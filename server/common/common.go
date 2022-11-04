package common

import (
	"Blog2Gin/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"runtime"
	"strings"
)

type modelFunc func() (map[string]any, error)

func ExecModel(group *errgroup.Group, dataChan chan map[string]any, f modelFunc) {
	group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				message := fmt.Sprintf("ExecModel: panic recovered: %s", r)
				err = fmt.Errorf(message)
				log.Error(trace(fmt.Sprintf("ExecModel: panic recovered: %s", r)))
			}
		}()

		ret, err := f()
		if err != nil {
			return err
		} else {
			dataChan <- ret
			return nil
		}
	})
}

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(4, pcs[:])

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

func BaseCount(g *errgroup.Group) (int, int, int, error) {
	dataChan := make(chan map[string]any, 3)
	ExecModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetCategoryCount()
			return map[string]any{"categoryCount": ret}, err
		})
	ExecModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetPostCount()
			return map[string]any{"totalCount": ret}, err
		})
	ExecModel(g, dataChan,
		func() (map[string]any, error) {
			ret, err := model.GetTagCount()
			return map[string]any{"tagCount": ret}, err
		})
	err := g.Wait()
	close(dataChan)
	if err != nil {
		return 0, 0, 0, err
	}
	dbData := map[string]any{}
	for data := range dataChan {
		for key := range data {
			dbData[key] = data[key]
		}
	}
	tagCount := dbData["tagCount"].(int)
	totalCount := dbData["totalCount"].(int)
	categoryCount := dbData["categoryCount"].(int)
	return tagCount, categoryCount, totalCount, nil
}

func PaginationData(indexCtx *IndexContextData, pageNum int, totalPages int) {
	pageRange := makeRange(1, totalPages)
	rightEnd := pageNum + 2
	if pageNum+2 > len(pageRange) {
		rightEnd = len(pageRange)
	}
	if pageNum == 1 {
		indexCtx.Right = pageRange[pageNum:rightEnd]
		if indexCtx.Right[len(indexCtx.Right)-1]+1 < totalPages {
			indexCtx.RightHasMore = true
		}
		if indexCtx.Right[len(indexCtx.Right)-1] < totalPages {
			indexCtx.Last = true
		}
	} else if pageNum == totalPages {
		var startNum int
		if (pageNum - 3) > 0 {
			startNum = pageNum - 3
		}
		indexCtx.Left = pageRange[startNum : pageNum-1]
		if indexCtx.Left[0] > 2 {
			indexCtx.LeftHasMore = true
		}
		if indexCtx.Left[0] > 1 {
			indexCtx.First = true
		}
	} else {
		var startNum int
		if (pageNum - 3) > 0 {
			startNum = pageNum - 3
		}
		indexCtx.Left = pageRange[startNum : pageNum-1]
		indexCtx.Right = pageRange[pageNum:rightEnd]
		if indexCtx.Right[len(indexCtx.Right)-1]+1 < totalPages {
			indexCtx.RightHasMore = true
		}
		if indexCtx.Right[len(indexCtx.Right)-1] < totalPages {
			indexCtx.Last = true
		}
		if indexCtx.Left[0] > 2 {
			indexCtx.LeftHasMore = true
		}
		if indexCtx.Left[0] > 1 {
			indexCtx.First = true
		}
	}
}

func makeRange(min, max int) []int {

	a := make([]int, max-min+1)

	for i := range a {
		a[i] = min + i
	}
	return a
}
