- ✅️ test plugin:TODO
- specific teach mode implementation
- test `if-then` in [Cond_2.josn](./interaction/Cond_2.json)
- test grammar templates by using common words
- 🚀implement template:S+V+P
- 🚀implement template: conj(if,and,or)+S+VerbPhrase
- implement template: what(when,how...)+be(can,do...)+S+verbPhrase
- extends all templates with adverbs
- 🚀implement pronoun-cacher
- 🚀implement pronoun-resolver
- implement third-single verb templates

### About S+V+P

#### Common Predicates

- ✅️V-link + adj(phrase)
- ✅️V-link + adj+preposition phrase
- ✅️V-link + adj+doing
- ✅️V-link + adj+to do
- ✅️V-link + noun(phrase)
- ✅️V-link + preposition phrase

### About Adverbial

- do sth+to do(purpose) ...
- do sth+conj ...
- V+doing sth
- do sth+preposition phrase
- do sth+adverb
- do sth+adverb+preposition phrase

### About preposition

- Prep+Object(noun(phrase),v-ing,clause)

### About Verb Phrase

- V+prep:give up,pass through

### Common Object

- noun(phrase)
- clause(e.g:what I want)
- pronoun
- gerund (phrase)
- to_do(e.g:want to ...;hope to ...)

### About Clause

- put `wh-`to the tail(what he gives you -> he gives you what)
- special case when the `wh-` serves as the subject (e.g: who helps him),don't transform

### About Wh- Ordered

- S+V+O -> wh+S+V
- S+IO+DO I give him a book -> wh+S+O/wh+S+O+prep
- S+O+OC I call him A -> wh+S+O