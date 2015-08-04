package com.tkido.tools

object Properties {
  import java.io.FileInputStream
  import java.util.Properties
  import scala.collection.JavaConverters._
  
  def apply(path:String) :Map[String, String] = {
    val prop = new Properties
    prop.load(new FileInputStream(path))
    prop.asScala.toMap
  }
  def apply[T](path:String, func:String => T) :Map[String, T] =
    apply(path).mapValues(func)
  
}