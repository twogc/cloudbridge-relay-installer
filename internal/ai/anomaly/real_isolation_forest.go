package anomaly

import (
	"math"
	"math/rand"
)

// NewRealIsolationForest создает новый Isolation Forest
func NewRealIsolationForest(nTrees int, contamination float64, featureCount int) *RealIsolationForest {
	forest := &RealIsolationForest{
		trees:        make([]*RealIsolationTree, nTrees),
		threshold:    contamination,
		featureCount: featureCount,
		maxSamples:   256,
	}

	// Создаем деревья
	for i := 0; i < nTrees; i++ {
		forest.trees[i] = &RealIsolationTree{
			heightLimit: int(math.Ceil(math.Log2(float64(forest.maxSamples)))),
			sampleSize:  forest.maxSamples,
		}
	}

	return forest
}

// Fit обучает лес на данных
func (rif *RealIsolationForest) Fit(samples [][]float64) {
	if len(samples) == 0 {
		return
	}

	// Обучаем каждое дерево на случайной подвыборке
	for _, tree := range rif.trees {
		if tree != nil {
			// Выбираем случайную подвыборку
			subsample := rif.getRandomSubsample(samples)
			tree.buildTree(subsample, 0)
		}
	}
}

// Predict предсказывает аномальность для данных
func (rif *RealIsolationForest) Predict(features []float64) (float64, error) {
	if len(rif.trees) == 0 {
		return 0.5, nil // По умолчанию средний скор
	}

	var totalScore float64
	validTrees := 0

	for _, tree := range rif.trees {
		if tree != nil && tree.root != nil {
			score := tree.predict(features)
			totalScore += score
			validTrees++
		}
	}

	if validTrees == 0 {
		return 0.5, nil
	}

	// Усредняем скора от всех деревьев
	avgScore := totalScore / float64(validTrees)

	// Нормализуем скор
	return math.Min(1.0, math.Max(0.0, avgScore)), nil
}

// getRandomSubsample получает случайную подвыборку
func (rif *RealIsolationForest) getRandomSubsample(samples [][]float64) [][]float64 {
	if len(samples) <= rif.maxSamples {
		return samples
	}

	// Выбираем случайные индексы
	indices := make([]int, rif.maxSamples)
	for i := range indices {
		indices[i] = rand.Intn(len(samples))
	}

	// Создаем подвыборку
	subsample := make([][]float64, rif.maxSamples)
	for i, idx := range indices {
		subsample[i] = make([]float64, len(samples[idx]))
		copy(subsample[i], samples[idx])
	}

	return subsample
}

// buildTree строит дерево изоляции
func (rit *RealIsolationTree) buildTree(samples [][]float64, height int) {
	if len(samples) <= 1 || height >= rit.heightLimit {
		// Создаем лист
		rit.root = &RealNode{
			isLeaf:       true,
			height:       height,
			sampleCount:  len(samples),
		}
		return
	}

	// Выбираем случайный признак
	featureCount := len(samples[0])
	if featureCount == 0 {
		rit.root = &RealNode{
			isLeaf:       true,
			height:       height,
			sampleCount:  len(samples),
		}
		return
	}

	feature := rand.Intn(featureCount)

	// Выбираем случайный порог
	minVal, maxVal := rit.getMinMax(samples, feature)
	if minVal == maxVal {
		// Все значения одинаковые, создаем лист
		rit.root = &RealNode{
			isLeaf:       true,
			height:       height,
			sampleCount:  len(samples),
		}
		return
	}

	threshold := minVal + rand.Float64()*(maxVal-minVal)

	// Разделяем данные
	leftSamples, rightSamples := rit.splitSamples(samples, feature, threshold)

	// Создаем узел
	rit.root = &RealNode{
		feature:     feature,
		threshold:   threshold,
		isLeaf:      false,
		height:      height,
		sampleCount: len(samples),
	}

	// Рекурсивно строим поддеревья
	if len(leftSamples) > 0 {
		rit.root.left = &RealNode{}
		leftTree := &RealIsolationTree{
			heightLimit: rit.heightLimit,
			sampleSize:  rit.sampleSize,
		}
		leftTree.buildTree(leftSamples, height+1)
		rit.root.left = leftTree.root
	}

	if len(rightSamples) > 0 {
		rit.root.right = &RealNode{}
		rightTree := &RealIsolationTree{
			heightLimit: rit.heightLimit,
			sampleSize:  rit.sampleSize,
		}
		rightTree.buildTree(rightSamples, height+1)
		rit.root.right = rightTree.root
	}
}

// getMinMax получает минимальное и максимальное значения для признака
func (rit *RealIsolationTree) getMinMax(samples [][]float64, feature int) (float64, float64) {
	if len(samples) == 0 {
		return 0, 0
	}

	minVal := samples[0][feature]
	maxVal := samples[0][feature]

	for _, sample := range samples {
		if feature < len(sample) {
			if sample[feature] < minVal {
				minVal = sample[feature]
			}
			if sample[feature] > maxVal {
				maxVal = sample[feature]
			}
		}
	}

	return minVal, maxVal
}

// splitSamples разделяет выборку по признаку и порогу
func (rit *RealIsolationTree) splitSamples(samples [][]float64, feature int, threshold float64) ([][]float64, [][]float64) {
	var leftSamples, rightSamples [][]float64

	for _, sample := range samples {
		if feature < len(sample) {
			if sample[feature] < threshold {
				leftSamples = append(leftSamples, sample)
			} else {
				rightSamples = append(rightSamples, sample)
			}
		}
	}

	return leftSamples, rightSamples
}

// predict предсказывает аномальность для точки
func (rit *RealIsolationTree) predict(features []float64) float64 {
	if rit.root == nil {
		return 0.5
	}

	pathLength := rit.getPathLength(rit.root, features, 0)
	
	// Нормализуем длину пути
	expectedPathLength := math.Log2(float64(rit.sampleSize))
	if expectedPathLength == 0 {
		return 0.5
	}

	// Вычисляем скор аномалии
	anomalyScore := math.Pow(2, -float64(pathLength)/expectedPathLength)
	
	return math.Min(1.0, math.Max(0.0, anomalyScore))
}

// getPathLength получает длину пути до изоляции точки
func (rit *RealIsolationTree) getPathLength(node *RealNode, features []float64, currentLength int) int {
	if node == nil {
		return currentLength
	}

	if node.isLeaf {
		return currentLength
	}

	if node.feature >= len(features) {
		return currentLength
	}

	// Переходим к следующему узлу
	if features[node.feature] < node.threshold {
		return rit.getPathLength(node.left, features, currentLength+1)
	} else {
		return rit.getPathLength(node.right, features, currentLength+1)
	}
}
