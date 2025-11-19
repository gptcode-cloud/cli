vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local M = require('chuchu')
M.setup({})

print("[TEST] Full flow test")

local mock_job_id = 12345
local stdin_received = {}

vim.fn.jobstart = function(cmd, opts)
  print("[TEST] jobstart called")
  
  vim.schedule(function()
    print("[TEST] Emitting events...")
    
    local events = {
      '__EVENT__{"type":"message","data":{"content":"Starting"}}__EVENT__',
      '__EVENT__{"type":"status","data":{"status":"Working"}}__EVENT__',
      '__EVENT__{"type":"confirm","data":{"prompt":"Continue?","id":"test1"}}__EVENT__',
    }
    
    for i, event in ipairs(events) do
      print(string.format("[TEST] Event %d/%d", i, #events))
      opts.on_stdout(mock_job_id, {event}, nil)
      vim.wait(50)
    end
    
    print("[TEST] Waiting for stdin response...")
    vim.wait(1000)
    
    print("[TEST] Received", #stdin_received, "stdin messages")
    for i, msg in ipairs(stdin_received) do
      print(string.format("[TEST] stdin[%d]: %s", i, msg))
    end
    
    opts.on_exit(mock_job_id, 0, nil)
  end)
  
  return mock_job_id
end

vim.fn.chansend = function(job, data)
  print("[TEST] chansend: job=" .. job .. " data=" .. data)
  table.insert(stdin_received, data)
  return 1
end

local input_called = false
vim.ui.input = function(opts, callback)
  print("[TEST] vim.ui.input: " .. opts.prompt)
  input_called = true
  vim.schedule(function()
    callback("y")
  end)
end

vim.schedule(function()
  M.toggle_chat()
  
  vim.wait(100)
  
  print("[TEST] Calling send_to_llm")
  M.send_to_llm("test message")
  
  vim.wait(3000)
  
  print("\n[TEST] Results:")
  print("  - vim.ui.input called:", input_called)
  print("  - stdin messages:", #stdin_received)
  
  vim.cmd('qall!')
end)
