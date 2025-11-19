package.loaded['chuchu'] = nil
vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local M = dofile('/Users/jadercorrea/workspace/opensource/chuchu/neovim/lua/chuchu/init.lua')
M.setup({})

print("=== DIAGNOSTIC TEST ===\n")

local stdout_received = false
local events_parsed = 0
local raw_lines = {}

local original_jobstart = vim.fn.jobstart
vim.fn.jobstart = function(cmd, opts)
  print("[1] Testing jobstart callback...")
  
  local job = original_jobstart(cmd, {
    cwd = opts.cwd,
    stdin = "pipe",
    stdout_buffered = false,
    on_stdout = function(j, data, n)
      if data then
        for _, line in ipairs(data) do
          if line ~= "" then
            stdout_received = true
            table.insert(raw_lines, line)
            print("[STDOUT] " .. line)
            
            local event_match = line:match("__EVENT__(.+)__EVENT__")
            if event_match then
              events_parsed = events_parsed + 1
              print("[PARSED] Event #" .. events_parsed)
            end
          end
        end
      end
      if opts.on_stdout then opts.on_stdout(j, data, n) end
    end,
    on_exit = function(j, c, n)
      print("\n=== RESULTS ===")
      print("Stdout received:", stdout_received)
      print("Raw lines:", #raw_lines)
      print("Events parsed:", events_parsed)
      
      if not stdout_received then
        print("\n❌ PROBLEM: jobstart callback not receiving stdout")
        print("   Check if stdout_buffered=false is working")
      elseif events_parsed == 0 then
        print("\n❌ PROBLEM: Events not being parsed")
        print("   First line:", raw_lines[1] and raw_lines[1]:sub(1, 80) or "none")
        print("   Check regex pattern: __EVENT__(.+)__EVENT__")
      else
        print("\n✅ Both working correctly")
      end
      
      if opts.on_exit then opts.on_exit(j, c, n) end
      vim.cmd('qall!')
    end,
  })
  
  return job
end

vim.fn.chansend = function(job, data)
  print("[2] Stdin sent: " .. data:sub(1, 100))
  return vim.fn.chansend(job, data)
end

vim.schedule(function()
  M.send_to_llm("Analyse codebase and add pix")
  vim.wait(8000)
  print("\n⏱ Timeout - forcing exit")
  vim.cmd('qall!')
end)
