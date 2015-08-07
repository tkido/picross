package com.tkido.picross

object Parser {
  import com.tkido.tools.Logger
  import com.tkido.tools.Text
  import com.tkido.tools.Math
  
  def apply(path:String) :Puzzle = {
    val lines = Text.readLines(path)
    Logger.debug(lines)
    val sizes = lines.head.split(SPLITTER).toList.map(_.toInt)
    
    Logger.debug(sizes)
    
    val hints = lines.tail.map(_.split(SPLITTER).toList.map(_.toInt))
    Logger.debug(hints)
    val hintsPair = hints.splitAt(sizes(0))
    Logger.debug(hintsPair)
    val hintsList = List(hintsPair._1, hintsPair._2)
    Logger.debug(hintsList)
    
    assert(hints.size == sizes.sum, "盤面の大きさとヒントの数が一致しません。")
    assert(hintsList(0).map(_.sum).sum == hintsList(1).map(_.sum).sum, "横のヒントの合計と縦のヒントの合計が一致しません。")
    for((hints, i) <- hintsList.zipWithIndex){
      for(hint <- hints){
        assert(sizes(1-i) >= hint.sum + hint.size - 1, "ヒントの合計が大きすぎる行または列があります。")
      }
    }
    
    val changes =
      for((hints, i) <- hintsList.zipWithIndex)
        yield hints.map(hint => hint.max >= sizes(1-i) - (hint.sum + hint.size - 1))
    Logger.debug(changes)
    
    val estimates =
      for((hints, i) <- hintsList.zipWithIndex)
        yield hints.map(hint => Math.combination(sizes(1-i) - hint.sum + 1, hint.size))
    Logger.debug(estimates)
    
    val hintsNew =
      for((hints, i) <- hintsList.zipWithIndex)
        yield hints.map(hint =>
          if(hint.sum == 0)
            List(sizes(1-i))
          else
            List(0) ::: hint.flatMap(List(_, 1)).dropRight(1) ::: List(0)
        )
    Logger.debug(hintsNew)
    
    
    
    new Puzzle()
  }
}