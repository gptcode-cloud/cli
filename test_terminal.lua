local result_file = '/tmp/nvim_test_result.txt'
local f = io.open(result_file, 'w')

local cmd = {"chu", "chat"}
local got_stdout = false
local events = {}

local job = vim.fn.jobstart(cmd, {
  stdin = "pipe",
  stdout_buffered = false,
  on_stdout = function(j, data, n)
    if data then
      for _, line in ipairs(data) do
        if line ~= '' then
          got_stdout = true
          table.insert(events, line)
          f:write('LINE: ' .. line .. '\n')
          f:flush()
        end
      end
    end
  end,
  on_exit = function(j, code, n)
    f:write('\nEXIT: ' .. code .. '\n')
    f:write('GOT_STDOUT: ' .. tostring(got_stdout) .. '\n')
    f:write('EVENTS: ' .. #events .. '\n')
    f:close()
    vim.cmd('qall!')
  end,
})

if job > 0 then
  vim.fn.chansend(job, '{"messages":[{"role":"user","content":"Analyse add pix"}]}\n')
else
  f:write('JOB FAILED\n')
  f:close()
  vim.cmd('qall!')
end

vim.defer_fn(function()
  f:write('\nTIMEOUT\n')
  f:close()
  vim.cmd('qall!')
end, 6000)
