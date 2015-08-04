package com.tkido

package object tools {
  def tryOrElse[T](function:() => T, default:T) :T = {
    try{
      function()
    }catch{
      case error:Throwable => {
        Logger.fatal(error)
        default
      }
    }
  }
}