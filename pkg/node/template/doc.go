// Package template provides a handler to get template nodes with a given one.
//   - special feature: if two placeholders come one after another,
//     then the first one always get only 1 word
//     E.g. I like eating some cakes -> I like $1{eating} $2{some cakes}
package template
