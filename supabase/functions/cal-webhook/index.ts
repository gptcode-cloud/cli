import "jsr:@supabase/functions-js/edge-runtime.d.ts"
import { createClient } from 'jsr:@supabase/supabase-js@2'

const WEBHOOK_SECRET = '05e6757489a1c040206f8d56b2248fae5ba525ea899e2c662a137b0079bcb330'

Deno.serve(async (req) => {
  try {
    const body = await req.text()
    const signature = req.headers.get('X-Cal-Signature-256')
    
    // Validate webhook signature (optional, Cal.com sends this)
    if (signature && WEBHOOK_SECRET) {
      const encoder = new TextEncoder()
      const key = await crypto.subtle.importKey(
        'raw',
        encoder.encode(WEBHOOK_SECRET),
        { name: 'HMAC', hash: 'SHA-256' },
        false,
        ['sign', 'verify']
      )
      const expectedSignature = await crypto.subtle.sign(
        'HMAC',
        key,
        encoder.encode(body)
      )
      const expectedHex = Array.from(new Uint8Array(expectedSignature))
        .map(b => b.toString(16).padStart(2, '0'))
        .join('')
      
      if (signature !== expectedHex) {
        console.error('Invalid webhook signature')
        return new Response(
          JSON.stringify({ error: 'Invalid signature' }),
          { status: 401, headers: { "Content-Type": "application/json" } }
        )
      }
    }

    const supabase = createClient(
      Deno.env.get('SUPABASE_URL') ?? '',
      Deno.env.get('SUPABASE_SERVICE_ROLE_KEY') ?? ''
    )

    const payload = JSON.parse(body)
    console.log('Cal.com webhook received:', payload)

    // Extract email from various possible locations
    const email = payload.responses?.email || 
                  payload.attendees?.[0]?.email || 
                  payload.responses?.email?.value

    if (!email) {
      console.error('No email found in payload')
      return new Response(
        JSON.stringify({ error: 'Email not found' }),
        { status: 400, headers: { "Content-Type": "application/json" } }
      )
    }

    // Determine product from metadata (query params) or fallback to event title
    const product = payload.metadata?.product || 
                    payload.responses?.product ||
                    (payload.eventType?.title?.toLowerCase()?.includes('cloud') ? 'cloud' : 'live')

    // Upsert into leads table - update if email exists, insert if new
    const { data, error } = await supabase
      .from('leads')
      .upsert({
        email: email,
        product: product,
        interview_scheduled: true,
        interview_slot: payload.startTime,
        cal_event_id: payload.uid,
        status: 'interview_scheduled',
      }, {
        onConflict: 'email',
        ignoreDuplicates: false
      })
      .select()

    if (error) {
      console.error('Supabase insert error:', error)
      return new Response(
        JSON.stringify({ error: error.message }),
        { status: 500, headers: { "Content-Type": "application/json" } }
      )
    }

    console.log('Lead updated with interview:', data)
    return new Response(
      JSON.stringify({ success: true, data }),
      { headers: { "Content-Type": "application/json" } }
    )
  } catch (err) {
    console.error('Webhook error:', err)
    return new Response(
      JSON.stringify({ error: err.message }),
      { status: 500, headers: { "Content-Type": "application/json" } }
    )
  }
})

/* To invoke locally:

  1. Run `supabase start` (see: https://supabase.com/docs/reference/cli/supabase-start)
  2. Make an HTTP request:

  curl -i --location --request POST 'http://127.0.0.1:54321/functions/v1/cal-webhook' \
    --header 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0' \
    --header 'Content-Type: application/json' \
    --data '{"name":"Functions"}'

*/
