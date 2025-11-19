vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

print("[TEST] Direct send_to_llm test")

local M = require('chuchu')
M.setup({})

local events_captured = {}

local original_jobstart = vim.fn.jobstart
vim.fn.jobstart = function(cmd, opts)
  print("[TEST] jobstart:", table.concat(cmd, " "))
  
  local job = original_jobstart(cmd, {
    cwd = opts.cwd,
    stdin = "pipe",
    stdout_buffered = false,
    on_stdout = function(j, data, n)
      for _, line in ipairs(data or {}) do
        if line ~= "" then
          print("[STDOUT]", line:sub(1, 100))
          table.insert(events_captured, line)
        end
      end
      if opts.on_stdout then opts.on_stdout(j, data, n) end
    end,
    on_exit = function(j, c, n)
      print("[EXIT]", c)
      if opts.on_exit then opts.on_exit(j, c, n) end
    end,
  })
  
  print("[TEST] Job ID:", job)
  return job
end

vim.ui.input = function(opts, cb)
  print("[INPUT]", opts.prompt)
  vim.schedule(function() cb("y") end)
end

vim.schedule(function()
  print("[TEST] Calling send_to_llm")
  M.send_to_llm("Analyse codebase and add pix")
  
  vim.wait(6000)
  
  print("\n[RESULTS]")
  print("Events:", #events_captured)
  for i, e in ipairs(events_captured) do
    print(i, e:sub(1, 80))
  end
  
  vim.cmd('qall!')
end)
