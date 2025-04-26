// Copyright 2018-20 PJ Engineering and Business Solutions Pty. Ltd. All rights reserved.

package hw

import (
	"context"
	"math"
)

type trainingState struct {
	initialSmooth        float64
	initialTrend         float64
	initialSeasonalComps []float64
	smoothingLevel       float64
	trendLevel           float64
	seasonalComps        []float64
	rmse                 float64
	T                    uint // how many observed values used in the forcasting process
}

func (hw *HoltWinters) trainSeries(ctx context.Context, start, end int) error {

	var (
		α, β, γ        float64 = hw.cfg.Alpha, hw.cfg.Beta, hw.cfg.Gamma
		period         int     = int(hw.cfg.Period)
		trnd, prevTrnd float64 // trend
		st, prevSt     float64 // smooth
	)

	y := hw.sf.Values[start : end+1]
	y_start := 0
	y_end := len(y) - 1

	seasonals := initialSeasonalComponents(y, period, hw.cfg.SeasonalMethod)

	hw.tstate.initialSeasonalComps = initialSeasonalComponents(y, period, hw.cfg.SeasonalMethod)

	trnd = initialTrend(y, period)
	hw.tstate.initialTrend = trnd

	var mse float64 // mean squared error

	// Training smoothing Level
	for i := y_start; i < y_end+1; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		xt := y[i]

		if i == y_start { // Set initial smooth
			st = xt
			hw.tstate.initialSmooth = xt
		} else {
			if hw.cfg.SeasonalMethod == Multiplicative {
				// multiplicative method
				prevSt, st = st, α*(xt/seasonals[i%period])+(1-α)*(st+trnd)
				trnd = β*(st-prevSt) + (1-β)*trnd
				seasonals[i%period] = γ*(xt/st) + (1-γ)*seasonals[i%period]
			} else {
				// additive method
				prevSt, st = st, α*(xt-seasonals[i%period])+(1-α)*(st+trnd)
				prevTrnd, trnd = trnd, β*(st-prevSt)+(1-β)*trnd
				seasonals[i%period] = γ*(xt-prevSt-prevTrnd) + (1-γ)*seasonals[i%period]
			}

			err := (xt - seasonals[i%period]) // actual value - smoothened value
			mse = mse + err*err
		}

	}
	hw.tstate.T = uint(end - start + 1)
	hw.tstate.rmse = math.Sqrt(mse / float64(end-start))

	hw.tstate.smoothingLevel = st
	hw.tstate.trendLevel = trnd
	hw.tstate.seasonalComps = seasonals

	return nil
}
