vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local M = require('chuchu')
M.setup({})

print("[TEST] Real flow simulation with actual chu binary")

local mock_job_id = 999
local job_started = false
local events_received = {}
local stdin_sent = {}

vim.fn.jobstart = function(cmd, opts)
  print("[TEST] jobstart called with:", vim.inspect(cmd))
  job_started = true
  
  local actual_job = vim.fn.jobstart(cmd, {
    cwd = opts.cwd,
    stdin = "pipe",
    stdout_buffered = false,
    on_stdout = function(job, data, name)
      if data then
        for _, line in ipairs(data) do
          if line ~= "" then
            print("[TEST] STDOUT:", line:sub(1, 120))
            local event_match = line:match("__EVENT__(.+)__EVENT__")
            if event_match then
              table.insert(events_received, event_match)
            end
          end
        end
      end
      if opts.on_stdout then
        opts.on_stdout(job, data, name)
      end
    end,
    on_stderr = function(job, data, name)
      if data then
        for _, line in ipairs(data) do
          if line ~= "" then
            print("[TEST] STDERR:", line:sub(1, 120))
          end
        end
      end
      if opts.on_stderr then
        opts.on_stderr(job, data, name)
      end
    end,
    on_exit = function(job, code, name)
      print("[TEST] EXIT code:", code)
      if opts.on_exit then
        opts.on_exit(job, code, name)
      end
    end,
  })
  
  print("[TEST] Real job id:", actual_job)
  return actual_job
end

local original_chansend = vim.fn.chansend
vim.fn.chansend = function(job, data)
  print("[TEST] chansend job=" .. job .. " len=" .. #data)
  table.insert(stdin_sent, data:sub(1, 100))
  return original_chansend(job, data)
end

local input_called = 0
vim.ui.input = function(opts, callback)
  input_called = input_called + 1
  print("[TEST] vim.ui.input #" .. input_called .. ": " .. opts.prompt)
  vim.schedule(function()
    print("[TEST] Sending response: y")
    callback("y")
  end)
end

vim.schedule(function()
  M.toggle_chat()
  vim.wait(100)
  
  print("[TEST] Sending message...")
  M.send_to_llm("Analyse codebase and add pix payment")
  
  vim.schedule(function()
    vim.wait(8000)
    
    print("\n==================== RESULTS ====================")
    print("Job started:", job_started)
    print("Events received:", #events_received)
    print("Stdin messages:", #stdin_sent)
    print("Input dialogs:", input_called)
    
    print("\nFirst 3 events:")
    for i = 1, math.min(3, #events_received) do
      local ev = events_received[i]
      print("  " .. i .. ":", ev:sub(1, 80))
    end
    
    print("\nStdin messages:")
    for i, msg in ipairs(stdin_sent) do
      print("  " .. i .. ":", msg)
    end
    
    if input_called == 0 then
      print("\n❌ PROBLEM: vim.ui.input never called!")
      print("Check if confirm events are being parsed")
    end
    
    if #events_received == 0 then
      print("\n❌ PROBLEM: No events received from stdout!")
      print("Check if chu binary is emitting events")
    end
    
    vim.cmd('qall!')
  end)
end)

vim.schedule(function()
  vim.wait(10000)
  print("[TEST] TIMEOUT")
  vim.cmd('qall!')
end)
