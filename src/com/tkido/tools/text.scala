package com.tkido.tools

object Text {
  import java.io.FileOutputStream
  import java.io.OutputStreamWriter
  import java.io.PrintWriter
  import scala.io.Source
  
  val reUrl = """^https?://.*""".r
  
  def readLines(source:String, charset:String = "utf-8") :List[String] = {
    source match {
      case reUrl() => readLinesFromUrl(source, charset)
      case _       => readLinesFromFile(source, charset)
    }
  }
  def read(source:String, charset:String = "utf-8") :String =
    readLines(source, charset).mkString("\n")
  
  private def readLinesFromUrl(url:String, charset:String) :List[String] =
    Source.fromURL(url, charset).getLines.toList.map(_.stripLineEnd)
  private def readLinesFromFile(path:String, charset:String) :List[String] = {
    val s = Source.fromFile(path, charset)
    val lines = try s.getLines.toList finally s.close
    lines.map(_.stripLineEnd)
  }
  
  def write(path:String, data:String, charset:String = "utf-8") {
    val fos = new FileOutputStream(path)
    val osw = new OutputStreamWriter(fos, charset)
    val pw  = new PrintWriter(osw)
    pw.println(data)
    pw.close
  }
  
}