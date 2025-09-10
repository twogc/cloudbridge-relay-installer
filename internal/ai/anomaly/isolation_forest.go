package anomaly

import (
	"math"
	"math/rand"
)

// SimpleIsolationForest простая реализация Isolation Forest
type SimpleIsolationForest struct {
	trees        []*IsolationTree
	threshold    float64
	featureCount int
}

// IsolationTree дерево изоляции
type IsolationTree struct {
	root        *Node
	heightLimit int
}

// Node узел дерева
type Node struct {
	left      *Node
	right     *Node
	feature   int
	threshold float64
	isLeaf    bool
	height    int
}

// NewSimpleIsolationForest создает новый простой Isolation Forest
func NewSimpleIsolationForest(nTrees int, contamination float64) *SimpleIsolationForest {
	return &SimpleIsolationForest{
		trees:     make([]*IsolationTree, nTrees),
		threshold: contamination,
	}
}

// Predict предсказывает аномальность для данных
func (sif *SimpleIsolationForest) Predict(features []float64) (float64, error) {
	if len(sif.trees) == 0 {
		return 0.5, nil // По умолчанию средний score
	}

	var totalScore float64
	for _, tree := range sif.trees {
		if tree != nil {
			score := tree.predict(features)
			totalScore += score
		}
	}

	// Средний score по всем деревьям
	averageScore := totalScore / float64(len(sif.trees))

	return averageScore, nil
}

// Train обучает лес на данных
func (sif *SimpleIsolationForest) Train(data [][]float64) error {
	if len(data) == 0 {
		return nil
	}

	sif.featureCount = len(data[0])

	// Создаем деревья
	for i := range sif.trees {
		sif.trees[i] = &IsolationTree{
			heightLimit: int(math.Ceil(math.Log2(float64(len(data))))),
		}
		sif.trees[i].buildTree(data)
	}

	return nil
}

// buildTree строит дерево изоляции
func (it *IsolationTree) buildTree(data [][]float64) {
	it.root = it.buildNode(data, 0)
}

// buildNode строит узел дерева
func (it *IsolationTree) buildNode(data [][]float64, height int) *Node {
	if height >= it.heightLimit || len(data) <= 1 {
		return &Node{
			isLeaf: true,
			height: height,
		}
	}

	// Выбираем случайную фичу
	featureCount := len(data[0])
	feature := rand.Intn(featureCount)

	// Находим min и max значения для фичи
	minVal := data[0][feature]
	maxVal := data[0][feature]

	for _, point := range data {
		if point[feature] < minVal {
			minVal = point[feature]
		}
		if point[feature] > maxVal {
			maxVal = point[feature]
		}
	}

	// Если все значения одинаковые, создаем лист
	if minVal == maxVal {
		return &Node{
			isLeaf: true,
			height: height,
		}
	}

	// Выбираем случайный порог
	threshold := minVal + rand.Float64()*(maxVal-minVal)

	// Разделяем данные
	var leftData, rightData [][]float64
	for _, point := range data {
		if point[feature] < threshold {
			leftData = append(leftData, point)
		} else {
			rightData = append(rightData, point)
		}
	}

	// Если одно из подмножеств пустое, создаем лист
	if len(leftData) == 0 || len(rightData) == 0 {
		return &Node{
			isLeaf: true,
			height: height,
		}
	}

	// Создаем внутренний узел
	node := &Node{
		feature:   feature,
		threshold: threshold,
		height:    height,
		isLeaf:    false,
	}

	// Рекурсивно строим дочерние узлы
	node.left = it.buildNode(leftData, height+1)
	node.right = it.buildNode(rightData, height+1)

	return node
}

// predict предсказывает аномальность для точки данных
func (it *IsolationTree) predict(features []float64) float64 {
	if it.root == nil {
		return 0.5
	}

	pathLength := it.getPathLength(it.root, features)

	// Нормализуем path length
	avgPathLength := math.Log2(float64(len(features)))
	score := math.Pow(2, -pathLength/avgPathLength)

	return score
}

// getPathLength вычисляет длину пути до листа
func (it *IsolationTree) getPathLength(node *Node, features []float64) float64 {
	if node.isLeaf {
		return float64(node.height)
	}

	if features[node.feature] < node.threshold {
		return it.getPathLength(node.left, features)
	} else {
		return it.getPathLength(node.right, features)
	}
}

