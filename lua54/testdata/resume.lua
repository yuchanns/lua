local co = coroutine.create(function() fib(10) end)
while coroutine.status(co) ~= "dead" do
  print(coroutine.resume(co))
end
