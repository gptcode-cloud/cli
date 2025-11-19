package.loaded['chuchu'] = nil
package.loaded['chuchu.init'] = nil

vim.cmd([[set runtimepath+=~/workspace/opensource/chuchu/neovim]])

local M = dofile('/Users/jadercorrea/workspace/opensource/chuchu/neovim/lua/chuchu/init.lua')
M.setup({})

print("[TEST] Config chat_cmd:", vim.inspect(M._test_get_config and M._test_get_config().chat_cmd or "N/A"))

local events = {}
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
          print("[STDOUT]", line)
          table.insert(events, line)
        end
      end
      if opts.on_stdout then opts.on_stdout(j, data, n) end
    end,
    on_exit = function(j, c, n)
      print("[EXIT]", c)
      print("[TOTAL EVENTS]", #events)
      for i, e in ipairs(events) do
        print(string.format("  %d: %s", i, e:sub(1, 100)))
      end
      if opts.on_exit then opts.on_exit(j, c, n) end
    end,
  })
  
  return job
end

local original_chansend = vim.fn.chansend
vim.fn.chansend = function(job, data)
  print("[STDIN] job=" .. job .. " data=" .. data:sub(1, 100))
  return original_chansend(job, data)
end

vim.ui.input = function(opts, cb)
  print("[INPUT]", opts.prompt)
  vim.schedule(function() cb("y") end)
end

vim.schedule(function()
  M.send_to_llm("Analyse codebase and add pix")
  vim.wait(6000)
  vim.cmd('qall!')
end)
