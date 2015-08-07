package com.tkido.picross

object main extends App {
  import com.tkido.picross.Config
  import com.tkido.tools.Logger
  import com.tkido.tools.Text
  
  import com.tkido.tools.Math
  
  Logger.level = Config.logLevel
  
  Parser("data/q_bad.txt")

  Logger.close()
}
