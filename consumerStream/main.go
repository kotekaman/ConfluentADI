package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	"log"
	"regexp"
)

var (
	brokers             = []string{"localhost:9092"}
	topic   goka.Stream = "xxxx3"
	group   goka.Group  = "xxxx3-group"
)

type user struct {
	Word   string
	Jumlah int
}

type datax struct{}

func (d *datax) Encode(value interface{}) (data []byte, err error) {
	if _, isUser := value.(*[]user); !isUser {
		return nil, fmt.Errorf("Codec requires value *user, got %T", value)
	}
	return json.Marshal(value)
}

func (d *datax) Decode(data []byte) (value interface{}, err error) {
	var (
		c []user
		_ error
	)
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling user: %v", err)
	}
	return &c, nil
}

// emits a single message and leave
// process messages until ctrl-c is pressed
func runProcessor() {
	// process callback is invoked for each message delivered from
	// "example-stream" topic.
	var datas, datas2, datas3 []user
	var datatemp *[]user
	var datad, datad2 user

	cb := func(ctx goka.Context, msg interface{}) {
		var regex, err = regexp.Compile(`\w+`)

		if err != nil {
			fmt.Println(err.Error())
		}

		var res1 = regex.FindAllString(msg.(string), -1)
		for _, word := range res1 {
			datad.Word = word
			datad.Jumlah = 1
			datas = append(datas, datad)
		}

		if val := ctx.Value(); val != nil {
			datatemp = ctx.Value().(*[]user)

			for _, wCount := range *datatemp {
				datad2.Word = wCount.Word
				datad2.Jumlah = wCount.Jumlah
				datas2 = append(datas2, datad2)
			}

			datas = append(datas, datas2...)
			fmt.Println(*datatemp)
			freq := make(map[string]int)
			for _, wCount := range datas {
				freq[wCount.Word] += wCount.Jumlah
				wCount.Jumlah = freq[wCount.Word]
			}

			for index, word := range freq {
				datad.Word = index
				datad.Jumlah = word
				datas3 = append(datas3, datad)
			}
			ctx.SetValue(&datas3)
			datas3 = nil
			datas = nil
			datas2 = nil
		} else {

			ctx.SetValue(&datas)
			datas = nil

		}

	}

	// Define a new processor group. The group defines all inputs, outputs, and
	// serialization formats. The group-table topic is "example-group-table".
	fmt.Println(new(codec.String))
	g := goka.DefineGroup(group,
		goka.Input(topic, new(codec.String), cb),
		goka.Persist(new(datax)),
	)

	p, err := goka.NewProcessor(brokers, g)
	if err != nil {
		log.Fatalf("error creating processor: %v", err)
	}
	ctxs := context.Background()

	r := p.Run(ctxs)
	fmt.Println(r.Error())

}

func main() { // emits one message and stops
	runProcessor() // press ctrl-c to stop
}
