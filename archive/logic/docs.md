## Docs
### `@`
1. `@caller A`:when running an operator(e.g:`[repo] GET;[safe++]`),it is used to limit the context
2. `A @ B -> C`:used in a node that serves as a callback(e.g:after `[repo] GET B @Caller A `,you get the value of the key and so you see it)

### `[Then]`
`[repo] [SET] [Then] A -> B`:used to save the result of a node,
it will be used like this below
```
[input] A
[input] [repo] {ctx} @ [Then] A -> [safe--] Count of [CondGroup] B @Caller A
[output] [safe--] Count of [CondGroup] B @Caller A
```

### `[CondGroup]`
see example above of `[Then]`,
A has the right to control the lifecycle of B,
when the count of B's `[CondGroup]` is 0, B will be activated
```
[input] [repo] Count of [CondGroup] B -> 0
[output] B
```
when to activate:when the last controller of B is activated and do the minus 1 that make the count to be equal to 0

