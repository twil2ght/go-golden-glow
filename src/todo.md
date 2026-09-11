1. pronoun的cache需要完善
2. 已经学习过的res-cond组合，下次碰到肯定不用在学习一遍而是增加一个learned hook
3. $apple和$happy本质一样，让$happy在跑一次variable:Var+VP毫无意义,需要记录variable系列的template，
4. 然后通过把所有的$happy相关的替换为$apple进行简化
5. 通过varietyLearning泛化过的可以直接删了，因为泛化过的就包括了当前情况
6. (method:这节去掉原本的cond:add和[res]，直接走[cond&res]，这样必然能达到目的)
7. 学习具体动词的具体用法后要删除不对的模板
8. 说出结果时。默认说No. 1组 补全别的cond:
9. what is the result that I ask you to check your name?
10. Susie: say : it is Susie 
11. Susie: if my name is Susie
12. 目前认为只有legal句子才需要variableProjection;因为非legal的句子的variable传递都是通过Cond%Res组直接定下的，不用担心匹配不上
13. 只有legal这种纯粹的外来的话需要特定规则来得到variable
14. new:反驳上述观点
15. if也需要重置end_learning
16. 不需要给res加variableprojection:因为这些不过是对应函数的入口，本身无意义，应该给cond里的[executor]加才对,这些才是货真价实的可以调用的函数
17. double_noun:让前面的noun具有adj词性
