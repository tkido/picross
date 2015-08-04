package com.tkido.picross

object main extends App {
  import com.tkido.picross.Config
  import com.tkido.tools.Logger
  import com.tkido.tools.Text
  
  Logger.level = Config.logLevel
  
  val grid = Grid(8, 8, 2)
  println(grid)

  Logger.close()
}
