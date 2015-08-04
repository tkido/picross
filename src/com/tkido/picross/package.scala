package com.tkido

package object picross {
  import scala.collection.immutable.Map
  
  val LIVE    = -1
  val UNKNOWN =  0
  val DEAD    =  1
  
  val table = Map(LIVE -> '■',
                  UNKNOWN -> '　',
                  DEAD -> '×')
}