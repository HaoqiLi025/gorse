package model

import "github.com/zhenghaoz/gorse/core"
import "github.com/zhenghaoz/gorse/base"

/* Base Model */

// BaseModel structure of all estimators.
type BaseModel struct {
	Params          base.Params          // Hyper-parameters
	UserIdSet       base.SparseIdSet     // Users' ID set
	ItemIdSet       base.SparseIdSet     // Items' ID set
	rng             base.RandomGenerator // Random generator
	randState       int64                // Random seed
	rtOptions       *base.FitOptions     // Runtime options
	getParamsCalled bool
}

func (model *BaseModel) SetParams(params base.Params) {
	model.getParamsCalled = true
	model.Params = params
	model.randState = model.Params.GetInt64(base.RandomState, 0)
}

func (model *BaseModel) GetParams() base.Params {
	return model.Params
}

func (model *BaseModel) Predict(userId, itemId int) float64 {
	panic("Predict() not implemented")
}

func (model *BaseModel) Fit(trainSet core.DataSet, options ...base.FitOption) {
	panic("Fit() not implemented")
}

// Init the base model.
func (model *BaseModel) Init(trainSet core.DataSet, options []base.FitOption) {
	// Check BaseModel.GetParams() called
	if model.getParamsCalled == false {
		panic("BaseModel.GetParams() not called")
	}
	// Setup ID set
	model.UserIdSet = trainSet.UserIdSet
	model.ItemIdSet = trainSet.ItemIdSet
	// Setup random state
	model.rng = base.NewRandomGenerator(model.randState)
	// Setup runtime options
	model.rtOptions = base.NewFitOptions(options)
}

/* Random */

// Random predicts a random rating based on the distribution of
// the training set, which is assumed to be normal. The prediction
// \hat{r}_{ui} is generated from a normal distribution N(\hat{μ},\hat{σ}^2)
// where \hat{μ} and \hat{σ}^2 are estimated from the training data
// using Maximum Likelihood Estimation
type Random struct {
	BaseModel
	// Parameters
	Mean   float64 // mu
	StdDev float64 // sigma
	Low    float64 // The lower bound of rating scores
	High   float64 // The upper bound of rating scores
}

// NewRandom creates a random model.
func NewRandom(params base.Params) *Random {
	random := new(Random)
	random.SetParams(params)
	return random
}

func (random *Random) Predict(userId int, itemId int) float64 {
	ret := random.rng.NormFloat64()*random.StdDev + random.Mean
	// Crop prediction
	if ret < random.Low {
		ret = random.Low
	} else if ret > random.High {
		ret = random.High
	}
	return ret
}

func (random *Random) Fit(trainSet core.DataSet, options ...base.FitOption) {
	random.Init(trainSet, options)
	random.Mean = trainSet.Mean()
	random.StdDev = trainSet.StdDev()
	random.Low, random.High = trainSet.Min(), trainSet.Max()
}

/* Baseline */

// BaseLine predicts the baseline estimate for given user and item.
//
//                   \hat{r}_{ui} = b_{ui} = μ + b_u + b_i
//
// If user u is unknown, then the Bias b_u is assumed to be zero. The same
// applies for item i with b_i.
type BaseLine struct {
	BaseModel
	UserBias   []float64 // b_u
	ItemBias   []float64 // b_i
	GlobalBias float64   // mu
	reg        float64
	lr         float64
	nEpochs    int
}

// NewBaseLine creates a baseline model. Parameters:
//	 Reg 		- The regularization parameter of the cost function that is
// 				  optimized. Default is 0.02.
//	 Lr 		- The learning rate of SGD. Default is 0.005.
//	 NEpochs	- The number of iteration of the SGD procedure. Default is 20.
func NewBaseLine(params base.Params) *BaseLine {
	baseLine := new(BaseLine)
	baseLine.SetParams(params)
	return baseLine
}

func (baseLine *BaseLine) SetParams(params base.Params) {
	baseLine.BaseModel.SetParams(params)
	// Setup parameters
	baseLine.reg = baseLine.Params.GetFloat64(base.Reg, 0.02)
	baseLine.lr = baseLine.Params.GetFloat64(base.Lr, 0.005)
	baseLine.nEpochs = baseLine.Params.GetInt(base.NEpochs, 20)
}

func (baseLine *BaseLine) Predict(userId, itemId int) float64 {
	denseUserId := baseLine.UserIdSet.ToDenseId(userId)
	denseItemId := baseLine.ItemIdSet.ToDenseId(itemId)
	return baseLine.predict(denseUserId, denseItemId)
}

func (baseLine *BaseLine) predict(denseUserId, denseItemId int) float64 {
	ret := baseLine.GlobalBias
	if denseUserId != base.NotId {
		ret += baseLine.UserBias[denseUserId]
	}
	if denseItemId != base.NotId {
		ret += baseLine.ItemBias[denseItemId]
	}
	return ret
}

func (baseLine *BaseLine) Fit(trainSet core.DataSet, options ...base.FitOption) {
	baseLine.Init(trainSet, options)
	// Initialize parameters
	baseLine.GlobalBias = trainSet.GlobalMean
	baseLine.UserBias = make([]float64, trainSet.UserCount())
	baseLine.ItemBias = make([]float64, trainSet.ItemCount())
	// Stochastic Gradient Descent
	for epoch := 0; epoch < baseLine.nEpochs; epoch++ {
		for i := 0; i < trainSet.Len(); i++ {
			denseUserId, denseItemId, rating := trainSet.GetDense(i)
			userBias := baseLine.UserBias[denseUserId]
			itemBias := baseLine.ItemBias[denseItemId]
			// Compute gradient
			diff := baseLine.predict(denseUserId, denseItemId) - rating
			gradUserBias := diff + baseLine.reg*userBias
			gradItemBias := diff + baseLine.reg*itemBias
			// Update parameters
			baseLine.UserBias[denseUserId] -= baseLine.lr * gradUserBias
			baseLine.ItemBias[denseItemId] -= baseLine.lr * gradItemBias
		}
	}
}

// ItemPop recommends items by their popularity.
type ItemPop struct {
	BaseModel
	Pop []float64
}

// NewItemPop creates an ItemPop model.
func NewItemPop(params base.Params) *ItemPop {
	pop := new(ItemPop)
	pop.SetParams(params)
	return pop
}

func (pop *ItemPop) Fit(set core.DataSet, options ...base.FitOption) {
	pop.Init(set, options)
	// Get items' popularity
	pop.Pop = make([]float64, set.ItemCount())
	for i := range set.DenseItemRatings {
		pop.Pop[i] = float64(set.DenseItemRatings[i].Len())
	}
}

func (pop *ItemPop) Predict(userId, itemId int) float64 {
	// Return items' popularity
	denseItemId := pop.ItemIdSet.ToDenseId(itemId)
	if denseItemId == base.NotId {
		return 0
	}
	return pop.Pop[denseItemId]
}
