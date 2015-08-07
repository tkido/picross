package com.tkido.tools

object Math {
  
  val memo:Array[BigInt] = Array()
  def factorial(n:Int) :BigInt = {
    if(n == 0)
      return 1
    else if(memo.isDefinedAt(n)){
      return memo(n)
    }else{
      val x = factorial(n - 1) * n
      memo :+ x
      return x
    }
  }
  
  def combination(m:Int, n:Int) :BigInt = {
    factorial(m) / (factorial(n) * factorial(m - n))
  }
  
  
}