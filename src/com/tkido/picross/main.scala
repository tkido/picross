package com.tkido.picross

object main extends App {
  import com.tkido.picross.Config
  import com.tkido.tools.Logger
  import com.tkido.tools.Text
  
  Logger.level = Config.logLevel
  
  Parser("data/sample.txt")

  Logger.close()
}
