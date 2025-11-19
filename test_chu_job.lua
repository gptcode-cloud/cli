local lines = {}
local job = vim.fn.jobstart({'chu', 'chat'}, {
  stdin = 'pipe',
  stdout_buffered = false,
  stderr_buffered = false,
  on_stdout = function(j, data)
    for _, line in ipairs(data or {}) do
      if line ~= '' then
        table.insert(lines, line)
        print('STDOUT:', line:sub(1, 100))
      end
    end
  end,
  on_stderr = function(j, data)
    for _, line in ipairs(data or {}) do
      if line ~= '' then
        table.insert(lines, line)
        print('STDERR:', line:sub(1, 100))
      end
    end
  end,
  on_exit = function(j, code)
    print('EXIT:', code, 'LINES:', #lines)
    vim.cmd('qall!')
  end,
})
print('JOB:', job)
if job > 0 then
  vim.fn.chansend(job, '{"messages":[{"role":"user","content":"Analyse add pix"}]}\n')
  print('STDIN SENT')
end
vim.defer_fn(function() 
  print('TIMEOUT')
  vim.cmd('qall!') 
end, 6000)
