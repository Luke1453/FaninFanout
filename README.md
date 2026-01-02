Closing in does not make the select pick the done case. A receive from a closed channel succeeds immediately with the zero value, so case taken <- <-in: will send that zero value. To stop when in closes, receive with the comma-ok and return if closed.

```
select {
case <-done:
  return
case v, ok := <-in:
  if !ok {
    return
  }
  taken <- v
}
```
