#!/usr/local/bin/ruby
#-*- coding: UTF-8 -*-

# Copyright: 2009 Takanori Kido
# url :http://tkido.com/mebius/
# mail:takanorikido@gmail.com

class Array
  def sum
    inject(0){|r, i| r + i}
  end
end
class Integer
  def factorial
    (1..self).inject(1){|r, i| r * i}
  end
  def combination(n)
    self.factorial / (n.factorial * (self - n).factorial)
  end
end

module Picross
  class Invalid < StandardError; end
  class Impossible < StandardError; end
  UNKNOWN = 0; OFF = 1; ON  = -1;
  TABLE = { UNKNOWN => '　', OFF => '×', ON => '■' }.freeze
  
  class Puzzle
    def initialize(lines, logging = false)
      @height, @width = lines.shift.split(/,|\s/).map!{|s| s.to_i}
      @grid = Array.new(@height){Array.new(@width, UNKNOWN)}
      @transposed = false
      @width_now = @width
      
      @estimates = []
      @changed = Array.new(@height + @width, false)
      @guess = 0
      @logging = logging
      
      @hints = lines.map do |line|
        line.split(/,|\s/).map!{|str| str.to_i}
      end
      raise Invalid, "盤面の大きさとヒントの数が一致しません。" unless @height + @width == @hints.size
      
      @r_hints = @hints[0, @height]
      @c_hints = @hints[@height, @width]
      raise Invalid, "横のヒントの合計と縦のヒントの合計が一致しません。" unless @r_hints.map{|hint| hint.sum}.sum == @c_hints.map{|hint| hint.sum}.sum
      
      hints_new = []
      each_line do |row, index, hint|
        raise Invalid, "ヒントの合計が大きすぎる行または列があります。" unless @width_now >= hint.sum + hint.size - 1
        if hint == [0]
          @changed[index] = true
          @estimates[index] = 1
          hints_new.push([@width_now])
        else
          @changed[index] = true if hint.max > @width_now - (hint.sum + hint.size - 1)
          @estimates[index] = (@width_now - hint.sum + 1).combination(hint.size)
          arr = [0]
          hint.each {|n| arr.push(n, 1)}
          arr.pop
          hints_new.push arr.push(0)
        end
      end
      @hints = hints_new
    end

    def to_s(cursor_row = nil)
      rh_max = @r_hints.map{|hint| hint.size }.max
      ch_max = @c_hints.map{|hint| hint.size }.max
      
      buf = []
      (ch_max-1).downto(0) do |n|
        str = '　' * rh_max + '｜'
        @c_hints.each do |hint|
          str += (hint[n]) ? sprintf("%2d", hint.reverse[n] % 100) : '　'
        end
        buf.push str + '｜'
      end
      buf.push '--' * rh_max + '＋' + '--' * @width_now + '＋'
      @r_hints.each_with_index do |hint, row|
        str = '　' * (rh_max - hint.size)
        hint.each{|n| str += sprintf("%2d", n % 100) }
        str += '｜'
        (0...@width_now).each{|col| str += TABLE[@grid[row][col]] }
        str += '｜'
        str += "<<<<" if row == cursor_row
        buf.push str
      end
      buf.push '--' * rh_max + '＋' + '--' * @width_now + '＋'
      buf.join("\n")
    end
    
    def dup
      copy = super
      @grid = Marshal.load(Marshal.dump(@grid))
      @changed = @changed.dup
      @estimates = @estimates.dup
      copy
    end
    
    def changed?
      @changed.include?(true)
    end
    def solved?
      not @grid.map{|row| row.include?(UNKNOWN)}.include?(true) and not changed?
    end
    
    def guess(val)
      loop do
        row = @guess / @width
        col = @guess % @width
        if @grid[row][col] == UNKNOWN
          self[row, col] = val
          @changed[row] = true
          return self
        end
        @guess += 1
      end
    end
    
    def scan
      while changed?
        each_line do |row, index, hint|
          if @changed[index]
            if @estimates[index] < @height * @width / 2
              solve(row, index, hint)
              @changed[index] = false
              puts "★\n" + to_s(row) if @logging
            else
              @estimates[index] = @estimates[index] * 4 / 5
            end
          end
        end
      end
      self
    end
    
    private
    def solve(row, index, hint)
      @answers = Array.new
      _solve(row, [], hint.dup, OFF)
      a_size = @answers.size
      raise Impossible if a_size == 0
      @answers.map! do |answer|
        arr = []
        answer.each_with_index{|num, i| arr.concat [(-1)**i] * num}
        arr
      end
      columns = @answers.transpose.map{|column| column.sum}
      columns.each_with_index do |sum, i|
        case sum
        when -a_size then self[row, i] = ON
        when  a_size then self[row, i] = OFF
        end
      end
    end
    
    def _solve(row, answer, hint, val)
      if hint.empty?
        @answers.push answer if answer.sum == @width_now
      elsif @grid[row][answer.sum, hint.first].include?(-val)
        return
      else
        if (answer.sum + hint.sum) < @width_now and val == OFF
          hint_dup = hint.dup
          hint_dup[0] += 1
          _solve(row, answer.dup, hint_dup, val)
        end
        answer.push hint.shift
        _solve(row, answer, hint, -val)
      end
    end
    
    def each_line
      @hints.each_with_index do |hint, index|
        row = @transposed ? index - @height : index
        yield row, index, hint
        transpose if index == @height - 1 or index == @height + @width - 1
      end
    end
    def transpose
      @grid = @grid.transpose
      @transposed = !@transposed
      @width_now = @transposed ? @height : @width
      @r_hints, @c_hints = @c_hints, @r_hints
    end
    
    def []=(row, col, val)
      if @grid[row][col] != val
        @grid[row][col] = val
        index = @transposed ? col : @height + col
        @changed[index] = true
        @estimates[index] = @estimates[index] * 4 / 5
      end
    end
  end
  
  class Solver
    def initialize(logging = false)
      @logging = logging
    end
    def solve(file)
      stack = Array.new
      puzzle = Puzzle.new(file.readlines, @logging)
      puts puzzle
      begin
        puzzle.scan
      rescue Impossible
        raise Invalid, "解がありません。"
      end
      until puzzle.solved?
        stack.push puzzle.dup
        begin
          puzzle.guess(ON).scan
        rescue Impossible
          puzzle = stack.pop
          begin
            puzzle.guess(OFF).scan
          rescue Impossible
            raise Invalid, "解がありません。" if stack.empty?
            puzzle = stack.pop
            puzzle.guess(OFF)
          end
        end
      end
      puts puzzle
    end
  end
  
  def self.main(filename, option)
    Solver.new(option == "-v").solve(File.open(filename, "r"))
  end
  
end

Picross.main(ARGV[0], ARGV[1])