package rbsa

import (
	. "github.com/badgerodon/lalg"
	"github.com/badgerodon/statistics"
	"github.com/badgerodon/quadprog"
	"os"
)

type (
	ReturnsBasedStyleAnalysis struct {
		indices []string
		returns map[string][]float64
	}
)

func New() *ReturnsBasedStyleAnalysis {
	return &ReturnsBasedStyleAnalysis{
		make([]string,0),
		make(map[string][]float64),
	}
}
func (this *ReturnsBasedStyleAnalysis) AddIndex(id string, returns []float64) {
	this.indices = append(this.indices, id)
	this.returns[id] = returns
}
func (this *ReturnsBasedStyleAnalysis) Run(returns []float64) (map[string]float64, os.Error) {
	if len(this.indices) == 0 {
		return nil, os.NewError("No indices were defined to run the analysis against")
	}
	this.returns["MAIN"] = returns
	
	// Build a matrix of all the index returns
	indexReturnsMatrix := this.getIndexReturnsMatrix()
	// Compute the covariance matrix of all the index returns
	covarianceMatrix := statistics.CovarianceMatrix(indexReturnsMatrix)
	// Extend the covariance matrix by adding 0s to the first row and column
	extendedCovarianceMatrix := this.getExtendedMatrix(covarianceMatrix)
	// Compute the variance for this item
	fundVariance := statistics.Variance(returns)
	// Set the first cell of the extended covariance matrix to 0
	extendedCovarianceMatrix.Set(0,0,fundVariance)
	
	// Make sure the covariance matrix is positive definite
	includedRows, fixedExtendedCovarianceMatrix := statistics.MakePositiveDefinite(extendedCovarianceMatrix)
	
	// Compute the covariance vector
	covarianceVector := this.getCovarianceVector(includedRows, "MAIN", fundVariance)
	
	// Create the constraint matrices
	constraintMatrix1 := this.getConstraintMatrix1(fixedExtendedCovarianceMatrix)
	constraintMatrix2 := this.getConstraintMatrix2(fixedExtendedCovarianceMatrix)
	// Create the constraint vectors
	constraintVector1 := this.getConstraintVector1(fixedExtendedCovarianceMatrix)
	constraintVector2 := this.getConstraintVector2(fixedExtendedCovarianceMatrix)
	
	// Find the solution
	solution, err := quadprog.Solve(fixedExtendedCovarianceMatrix, 
		covarianceVector, 
		constraintMatrix1, 
		constraintVector1, 
		constraintMatrix2, 
		constraintVector2,
	)
	
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]float64)
	
	for _, i := range this.indices {
		result[i] = 0
	}
	
	for i := 1; i < len(includedRows); i++ {
		result[this.indices[includedRows[i-1]]] = solution[i]
	}
	
	return result, nil
}
func (this *ReturnsBasedStyleAnalysis) getCovarianceVector(rows []int, item string, variance float64) Vector {
	return nil
}
func (this *ReturnsBasedStyleAnalysis) getConstraintMatrix1(mat Matrix) Matrix {
	return mat
}
func (this *ReturnsBasedStyleAnalysis) getConstraintMatrix2(mat Matrix) Matrix {
	return mat
}
func (this *ReturnsBasedStyleAnalysis) getConstraintVector1(mat Matrix) Vector {
	return nil
}
func (this *ReturnsBasedStyleAnalysis) getConstraintVector2(mat Matrix) Vector {
	return nil
}

func (this *ReturnsBasedStyleAnalysis) getExtendedMatrix(mat Matrix) Matrix {
	n := NewMatrix(mat.Rows+1, mat.Cols+1)
	for i := 0; i < n.Rows; i++ {
		for j := 0; j < n.Cols; j++ {
			if i == 0 || j == 0 {
				n.Set(i,j,0)
			} else {
				n.Set(i,j,mat.Get(i-1,j-1))
			}
		}
	}
	return n
}
func (this *ReturnsBasedStyleAnalysis) getIndexReturnsMatrix() Matrix {
	if len(this.indices) == 0 {
		return NewMatrix(0, 0)
	}
	sz := len(this.returns[this.indices[0]])
	n := NewMatrix(len(this.indices), sz)
	for i, key := range this.indices {
		rs := this.returns[key]
		for j, v := range rs {
			n.Set(i, j, v)
		}
	}
	return n
}