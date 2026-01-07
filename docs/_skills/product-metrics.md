---
layout: skill
title: Product Metrics
name: product-metrics
category: product
description: Implement proper analytics, tracking pixels, and conversion funnels to measure product success.
---

# Product Metrics

Guidelines for implementing analytics and tracking that provides actionable insights.

## When to Activate

- Adding analytics to a product
- Implementing conversion tracking
- Setting up event tracking
- Configuring marketing pixels

## Event Naming Conventions

### Use object_action format

```javascript
// GOOD - consistent, queryable
analytics.track('button_clicked', { button_name: 'signup' });
analytics.track('form_submitted', { form_name: 'checkout' });
analytics.track('page_viewed', { page_name: 'pricing' });
analytics.track('product_added_to_cart', { product_id: '123' });

// BAD - inconsistent
analytics.track('clickedButton');
analytics.track('Submit Form');
analytics.track('user saw pricing page');
```

### Include context in every event

```javascript
// GOOD - rich context
analytics.track('purchase_completed', {
  // What
  order_id: 'ord_123',
  products: ['prod_1', 'prod_2'],
  total: 99.99,
  currency: 'BRL',
  
  // Where
  page: '/checkout',
  source: 'web',
  
  // Who (enrich server-side)
  user_id: 'usr_456',
  
  // When (auto-added by SDK)
  timestamp: '2026-01-07T10:00:00Z',
  
  // Attribution
  utm_source: 'google',
  utm_campaign: 'winter_sale',
});
```

## Google Analytics 4 (GA4)

### Custom events

```javascript
// gtag.js
gtag('event', 'sign_up', {
  method: 'email',
  plan: 'pro',
});

gtag('event', 'purchase', {
  transaction_id: 'T12345',
  value: 99.99,
  currency: 'BRL',
  items: [{
    item_id: 'SKU_123',
    item_name: 'Premium Plan',
    price: 99.99,
    quantity: 1,
  }],
});
```

### Page views with custom dimensions

```javascript
gtag('config', 'G-XXXXXXX', {
  page_title: 'Pricing Page',
  page_path: '/pricing',
  user_id: 'usr_123', // Logged-in user
  custom_map: {
    dimension1: 'user_plan',
    dimension2: 'experiment_variant',
  },
  user_plan: 'free',
  experiment_variant: 'control',
});
```

## UTM Parameters

### Standard parameters

```
https://example.com/landing?
  utm_source=google          # Traffic source
  utm_medium=cpc             # Marketing medium
  utm_campaign=winter_sale   # Campaign name
  utm_term=ai+tools          # Paid keywords
  utm_content=banner_v2      # A/B test variant
```

### Persist across navigation

```javascript
// Store UTMs on first visit
function captureUTMs() {
  const params = new URLSearchParams(window.location.search);
  const utms = {};
  
  ['utm_source', 'utm_medium', 'utm_campaign', 'utm_term', 'utm_content']
    .forEach(key => {
      if (params.has(key)) {
        utms[key] = params.get(key);
      }
    });
  
  if (Object.keys(utms).length > 0) {
    sessionStorage.setItem('utms', JSON.stringify(utms));
  }
}

// Include in all events
function getTrackingContext() {
  const utms = JSON.parse(sessionStorage.getItem('utms') || '{}');
  return {
    ...utms,
    referrer: document.referrer,
    landing_page: sessionStorage.getItem('landing_page'),
  };
}
```

## Conversion Funnels

### Define clear funnel stages

```javascript
const FUNNEL_STAGES = {
  LANDING: 'funnel_landing_viewed',
  SIGNUP_STARTED: 'funnel_signup_started',
  SIGNUP_COMPLETED: 'funnel_signup_completed',
  TRIAL_STARTED: 'funnel_trial_started',
  PAYMENT_STARTED: 'funnel_payment_started',
  PURCHASE_COMPLETED: 'funnel_purchase_completed',
};

function trackFunnelStage(stage, metadata = {}) {
  analytics.track(stage, {
    funnel_name: 'main_conversion',
    funnel_step: Object.keys(FUNNEL_STAGES).indexOf(stage) + 1,
    ...metadata,
    ...getTrackingContext(),
  });
}
```

### Track drop-off points

```javascript
// Track where users abandon
window.addEventListener('beforeunload', () => {
  const currentStage = getCurrentFunnelStage();
  if (currentStage && !isConversionComplete()) {
    navigator.sendBeacon('/api/analytics', JSON.stringify({
      event: 'funnel_abandoned',
      stage: currentStage,
      time_on_page: getTimeOnPage(),
    }));
  }
});
```

## Marketing Pixels

### Facebook Pixel

```html
<!-- Base pixel -->
<script>
!function(f,b,e,v,n,t,s){...}(window,document,'script',
'https://connect.facebook.net/en_US/fbevents.js');
fbq('init', 'YOUR_PIXEL_ID');
fbq('track', 'PageView');
</script>

<!-- Conversion events -->
<script>
fbq('track', 'Lead', { content_name: 'trial_signup' });
fbq('track', 'Purchase', { value: 99.99, currency: 'BRL' });
</script>
```

### Google Ads Conversion

```javascript
gtag('event', 'conversion', {
  send_to: 'AW-XXXXXXX/YYYYYYY',
  value: 99.99,
  currency: 'BRL',
  transaction_id: 'T12345',
});
```

## Core Web Vitals

### Monitor performance metrics

```javascript
import { onCLS, onFID, onLCP } from 'web-vitals';

function sendToAnalytics({ name, delta, id }) {
  gtag('event', name, {
    event_category: 'Web Vitals',
    event_label: id,
    value: Math.round(name === 'CLS' ? delta * 1000 : delta),
    non_interaction: true,
  });
}

onCLS(sendToAnalytics);
onFID(sendToAnalytics);
onLCP(sendToAnalytics);
```

### Set performance budgets

```javascript
// Alert on poor performance
const BUDGETS = {
  LCP: 2500,  // 2.5s
  FID: 100,   // 100ms
  CLS: 0.1,   // 0.1
};

function checkBudget({ name, value }) {
  if (value > BUDGETS[name]) {
    console.warn(`${name} exceeded budget: ${value} > ${BUDGETS[name]}`);
    analytics.track('performance_budget_exceeded', {
      metric: name,
      value,
      budget: BUDGETS[name],
      page: window.location.pathname,
    });
  }
}
```

## Privacy & Compliance

### Respect consent

```javascript
// Only track after consent
function initAnalytics() {
  if (!hasUserConsent()) {
    return;
  }
  
  // Initialize tracking
  gtag('consent', 'update', {
    analytics_storage: 'granted',
    ad_storage: 'granted',
  });
}

// Provide opt-out
function optOutAnalytics() {
  window['ga-disable-G-XXXXXXX'] = true;
  localStorage.setItem('analytics_opt_out', 'true');
}
```
