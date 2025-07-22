local co = coroutine.create(function()
	fib(10)
end)
while coroutine.status(co) ~= "dead" do
	local _, yield = coroutine.resume(co)
	print("resume from lua: " .. yield)
end
