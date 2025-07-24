local co = coroutine.create(function()
  for i= 1, 2 do
    coroutine.yield(i)
  end
  return 99
end)

return co
