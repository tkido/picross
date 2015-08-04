package com.tkido.tools

class Tengine(path:String) {
  val re = """#\{(.*?)\}""".r
  
  val lines = Text.readLines(path)
  
  def apply(data:Map[String, String]) :String = {
    lines.map(line =>
      re.replaceAllIn(line, m => data.getOrElse(m.group(1), ""))
    ).mkString("\n")
  }
}

object Tengine {
  def apply(path:String) = new Tengine(path)
}