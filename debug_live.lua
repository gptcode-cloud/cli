local logfile = io.open("/tmp/gptcode_debug.log", "w")

local function log(msg)
  if logfile then
    logfile:write(os.date("%H:%M:%S") .. " " .. msg .. "\n")
    logfile:flush()
  end
  print("[DEBUG] " .. msg)
end

log("=== DEBUG START ===")

local original_jobstart = vim.fn.jobstart
vim.fn.jobstart = function(cmd, opts)
  log("jobstart: " .. table.concat(cmd, " "))
  
  local wrapped_opts = vim.deepcopy(opts)
  
  local original_stdout = opts.on_stdout
  wrapped_opts.on_stdout = function(j, data, n)
    if data then
      for _, line in ipairs(data) do
        if line ~= "" then
          log("STDOUT[" .. j .. "]: " .. line:sub(1, 150))
        end
      end
    end
    if original_stdout then
      original_stdout(j, data, n)
    end
  end
  
  local original_exit = opts.on_exit
  wrapped_opts.on_exit = function(j, c, n)
    log("EXIT[" .. j .. "]: code=" .. c)
    if original_exit then
      original_exit(j, c, n)
    end
  end
  
  local job = original_jobstart(cmd, wrapped_opts)
  log("Job ID: " .. job)
  return job
end

local original_chansend = vim.fn.chansend
vim.fn.chansend = function(job, data)
  log("chansend[" .. job .. "]: " .. data:sub(1, 100))
  return original_chansend(job, data)
end

local original_input = vim.ui.input
vim.ui.input = function(opts, callback)
  log("vim.ui.input: " .. opts.prompt)
  return original_input(opts, function(response)
    log("vim.ui.input response: " .. tostring(response))
    callback(response)
  end)
end

log("Hooks installed. Tail -f /tmp/gptcode_debug.log to monitor")
vim.notify("Debug hooks active. Check /tmp/gptcode_debug.log", vim.log.levels.INFO)
