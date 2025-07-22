local co = coroutine.create(function()
	fib(10)
end)
local is_loop = true
while is_loop do
	local status, res = coroutine.resume(co)
	if not status then
		is_loop = false
		return
	end
	if res == nil then
		is_loop = false
	end
end

-- FIXME: This is a hack to make sure the program exits cleanly.
-- otherwise, it raises an error back to Go: fatal error: exitsyscall: syscall frame is no longer valid
os.exit(0)
