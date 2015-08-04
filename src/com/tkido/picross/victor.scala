package com.tkido.picross

case class Vector(x:Int, y:Int) extends Ordered[Vector]{
  def unary_+ :Vector = Vector(x, y)
  def unary_- :Vector = Vector(-x, -y)
  
  def +(that:Vector) :Vector =
    Vector(x + that.x, y + that.y)
  def -(that:Vector) :Vector =
    Vector(x - that.x, y - that.y)

  def *(that: Vector) :Vector =
    Vector(x * that.x - y * that.y,
           x * that.y + y * that.x)
  def *(scalar:Int) :Vector =
    Vector(x * scalar, y * scalar)
  def /(scalar:Int) :Vector =
    Vector(x / scalar, y / scalar)
  
  def compare(that:Vector) =
    if(y != that.y) y - that.y
    else if(x != that.x) x - that.x
    else 0
  
  override def toString = "(%d,%d)".format(x, y)
}