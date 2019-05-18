package main

import (
	"fmt"
)

func main() {
	grid := [][]byte{{'1','1','0','0','0'},{'1','1','0','0','0'},{'0','0','1','0','0'},{'0','0','0','1','1'}}
	nums := numIslands(grid)
	fmt.Println(nums)
}
func numIslands(grid [][]byte) int {
	height := len(grid)
	if height == 0{
		return 0
	}
	width := len(grid[0])
	if width == 0{
		return 0
	}

	flag := [][]int{{1,0},{0,1},{-1,0},{0,-1}}

	sum := 0
	for i := 0; i < height; i++{
		for j := 0; j < width; j++{
			if grid[i][j] == '1'{
				sum++
				dfs(grid,i,j,height,width,flag)
			}
		}
	}
	return sum
}

func dfs(grid [][]byte,i,j,h,w int,flag [][]int)  {
	grid[i][j] = '0'
	for _,v := range flag{
		x,y := i+v[0],j+v[1]
		if safe(x,y,h,w) && grid[x][y] == '1'{
			dfs(grid,x,y,h,w,flag)
		}
	}
}


func safe(i,j,h,w int) bool {
	if i >= 0 && i < h && j >= 0 && j < w{
		return true
	}
	return false
}