可以通过模糊查询或者精确查询成语的名字、解释、典故、例子，调用的成语api请换成你自己的。
先执行DownIdiomJson的进行所需的成语下载.
例如:  ./downIdiom -keyword 马 -rows 10

然后在执行PrintIdiomInfo进行查询。
模糊查询
例如:  ./idiom -model ambiguous -keyword 马        

精确查询
例如:  ./idiom -model accurate -keyword 持戈试马