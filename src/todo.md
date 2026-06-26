- ✅️ test plugin:TODO
- specific teach mode implementation
- test `if-then` in [Cond_2.josn](interaction/Cond_2_if.json)
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

### New Todo

- implement the verify of system KV
- implement the low-level api of Speak
- implement the parse of system variables:reuse the logic of pronoun(make `special types of pronouns`:)
- currently,the special pronoun is fed by the user;So how to achieve the real-world usage?(no words like $Susie but just Susie)


#### Special pronouns:

- ✅️ object-like: can reuse pronoun-resolver completely but need small fix(it can add attributes)
- verb-like: need extra works
- adj-like: need extra works

#### real-world usage of $X

- find the word that appears twice or more,surely,this is stupid,So then Let Susie ask back by using other value
  (Susie->A), and let user tell if she is right(e.g: yes,you have understood the approach)

### TODO of Future

- when taught to say something,check if it is copyable(just say what he says), or you have to make a simple transformation
- then tell something,check if you should tell indirectly or directly

### TODO now

- some words can be tricky like (.. of ..),because those words can be used in real sentence,
how can you distinguish whether it is keyword or a normal word?