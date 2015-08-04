package com.tkido.picross

object Parser {
  import com.tkido.tools.Logger
  import com.tkido.tools.Text
  
  def apply(path:String) :Puzzle = {
    val lines = Text.readLines(path)
    Logger.debug(lines)
    val sizes = lines.head.split(SPLITTER).toList.map(_.toInt)
    
    Logger.debug(sizes)
    
    val hints = lines.tail.map(_.split(SPLITTER).toList.map(_.toInt))
    Logger.debug(hints)
    
    assert(hints.size == sizes.sum)
    assert(hints.take(sizes(0)).map(_.sum).sum == hints.takeRight(sizes(1)).map(_.sum).sum)
    
    val new_hints = hints.map(hint =>
      if(hint.sum == 0)
        List(8) //TODO
      else
        List(0) ::: hint.flatMap(List(_, 1)).dropRight(1) ::: List(0)
    )
    Logger.debug(new_hints)
    new Puzzle()
  }
}