return function(x)
  coroutine.yield(x, x * x)
  return x + x
end
