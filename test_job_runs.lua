local started = false
local exited = false
local exit_code = nil

local job = vim.fn.jobstart({'chu', 'chat'}, {
  stdin = 'pipe',
  on_exit = function(j, code)
    exited = true
    exit_code = code
    print('[EXIT]', code)
  end,
})

print('[JOB]', job)
started = job > 0

if started then
  vim.fn.chansend(job, '{"messages":[{"role":"user","content":"test"}]}\n')
  print('[STDIN] sent')
end

vim.defer_fn(function()
  print('[RESULT] started=' .. tostring(started) .. ' exited=' .. tostring(exited) .. ' code=' .. tostring(exit_code))
  vim.cmd('qall!')
end, 3000)
