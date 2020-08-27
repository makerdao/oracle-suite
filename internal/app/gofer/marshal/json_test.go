//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package marshal

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/makerdao/gofer/pkg/aggregator"
	"github.com/makerdao/gofer/pkg/model"
)

func TestJSONForPriceAggregate(t *testing.T) {
	pa := model.NewPriceAggregate(
		"agg",
		&model.PricePoint{
			Price: 1.0,
			Pair:  &model.Pair{Base: "a", Quote: "b"},
		},
		model.NewPriceAggregate(
			"med",
			&model.PricePoint{
				Price: 2.0,
				Pair:  &model.Pair{Base: "c", Quote: "d"},
			},
		),
	)

	j := newJSON(false)
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	assert.JSONEq(
		t,
		`
			[
			   {
				  "type":"agg",
				  "pair":{
					 "base":"a",
					 "quote":"b"
				  },
				  "price":1,
				  "prices":[
					 {
						"type":"med",
						"pair":{
						   "base":"c",
						   "quote":"d"
						},
						"price":2
					 }
				  ]
			   },
			   {
				  "type":"agg",
				  "pair":{
					 "base":"a",
					 "quote":"b"
				  },
				  "price":1,
				  "prices":[
					 {
						"type":"med",
						"pair":{
						   "base":"c",
						   "quote":"d"
						},
						"price":2
					 }
				  ]
			   }
			]
		`,
		string(r),
	)
}

func TestJSONForPriceAggregate_Async(t *testing.T) {
	pa := model.NewPriceAggregate(
		"agg",
		&model.PricePoint{
			Price: 1.0,
			Pair:  &model.Pair{Base: "a", Quote: "b"},
		},
		model.NewPriceAggregate(
			"med",
			&model.PricePoint{
				Price: 2.0,
				Pair:  &model.Pair{Base: "c", Quote: "d"},
			},
		),
	)

	j := newJSON(true)
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Write(pa, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	l := strings.Split(string(r), "\n")

	assert.Nil(t, err)

	expected := `
	   {
		  "type":"agg",
		  "pair":{
			 "base":"a",
			 "quote":"b"
		  },
		  "price":1,
		  "prices":[
			 {
				"type":"med",
				"pair":{
				   "base":"c",
				   "quote":"d"
				},
				"price":2
			 }
		  ]
	   }
	`

	assert.Len(t, l, 3)
	assert.JSONEq(t, expected, l[0])
	assert.JSONEq(t, expected, l[1])
	assert.Equal(t, "", l[2]) // because last json must be also followed by a new line
}

func TestJSONForExchange(t *testing.T) {
	ex := &model.Exchange{
		Name:   "foobar",
		Config: map[string]string{"foo": "bar"},
	}

	j := newJSON(false)
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	assert.JSONEq(t, `["foobar","foobar"]`, string(r))
}

func TestJSONForExchange_Async(t *testing.T) {
	ex := &model.Exchange{
		Name:   "foobar",
		Config: map[string]string{"foo": "bar"},
	}

	j := newJSON(true)
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Write(ex, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	l := strings.Split(string(r), "\n")
	assert.Nil(t, err)

	assert.Len(t, l, 3)
	assert.JSONEq(t, `"foobar"`, l[0])
	assert.JSONEq(t, `"foobar"`, l[1])
	assert.Equal(t, "", l[2]) // because last json must be also followed by a new line
}

func TestJSONForPriceModelMap(t *testing.T) {
	pmm := aggregator.PriceModelMap{
		aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}}: aggregator.PriceModel{
			Method:     "method",
			MinSources: 2,
			Sources: []aggregator.PriceRefPath{
				[]aggregator.PriceRef{{
					Origin: "origin",
					Pair:   aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}},
				}},
			},
		},
	}

	j := newJSON(false)
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	assert.Nil(t, err)

	expected := `
		[
		   {
			  "Pair":"A/B",
			  "Model":{
				 "method":"method",
				 "minSourceSuccess":2,
				 "sources":[
					[
					   {
						  "origin":"origin",
						  "pair":"A/B"
					   }
					]
				 ]
			  }
		   },
		   {
			  "Pair":"A/B",
			  "Model":{
				 "method":"method",
				 "minSourceSuccess":2,
				 "sources":[
					[
					   {
						  "origin":"origin",
						  "pair":"A/B"
					   }
					]
				 ]
			  }
		   }
		]
	`

	assert.JSONEq(t, expected, string(r))
}

func TestJSONForPriceModelMa_Async(t *testing.T) {
	pmm := aggregator.PriceModelMap{
		aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}}: aggregator.PriceModel{
			Method:     "method",
			MinSources: 2,
			Sources: []aggregator.PriceRefPath{
				[]aggregator.PriceRef{{
					Origin: "origin",
					Pair:   aggregator.Pair{Pair: model.Pair{Base: "A", Quote: "B"}},
				}},
			},
		},
	}

	j := newJSON(true)
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Write(pmm, nil))
	assert.Nil(t, j.Close())

	r, err := ioutil.ReadAll(j)
	l := strings.Split(string(r), "\n")
	assert.Nil(t, err)

	expected := `	
	   {
		  "Pair":"A/B",
		  "Model":{
			 "method":"method",
			 "minSourceSuccess":2,
			 "sources":[
				[
				   {
					  "origin":"origin",
					  "pair":"A/B"
				   }
				]
			 ]
		  }
	   }
	`

	assert.Len(t, l, 3)
	assert.JSONEq(t, expected, l[0])
	assert.JSONEq(t, expected, l[1])
	assert.Equal(t, "", l[2]) // because last json must be also followed by a new line
}
