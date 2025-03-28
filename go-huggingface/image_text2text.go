package huggingface

import (
	"context"
	"encoding/json"
	"errors"
)

type ImageText2TextParameters struct {
	// (Default: None). Integer to define the top tokens considered within the sample operation to create new text.
	TopK *int `json:"top_k,omitempty"`

	// (Default: None). Float to define the tokens that are within the sample` operation of text generation. Add
	// tokens in the sample for more probable to least probable until the sum of the probabilities is greater
	// than top_p.
	TopP *float64 `json:"top_p,omitempty"`

	// (Default: 1.0). Float (0.0-100.0). The temperature of the sampling operation. 1 means regular sampling,
	// 0 means top_k=1, 100.0 is getting closer to uniform probability.
	Temperature *float64 `json:"temperature,omitempty"`

	// (Default: None). Float (0.0-100.0). The more a token is used within generation the more it is penalized
	// to not be picked in successive generation passes.
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"`

	// (Default: None). Int (0-250). The amount of new tokens to be generated, this does not include the input
	// length it is a estimate of the size of generated text you want. Each new tokens slows down the request,
	// so look for balance between response times and length of text generated.
	MaxNewTokens *int `json:"max_new_tokens,omitempty"`

	// (Default: None). Float (0-120.0). The amount of time in seconds that the query should take maximum.
	// Network can cause some overhead so it will be a soft limit. Use that in combination with max_new_tokens
	// for best results.
	MaxTime *float64 `json:"max_time,omitempty"`

	// (Default: True). Bool. If set to False, the return results will not contain the original query making it
	// easier for prompting.
	ReturnFullText *bool `json:"return_full_text,omitempty"`

	// (Default: 1). Integer. The number of proposition you want to be returned.
	NumReturnSequences *int `json:"num_return_sequences,omitempty"`
}

type ImageText2TextRequest struct {
	// String to generated from
	Inputs     string                   `json:"inputs"`
	Parameters ImageText2TextParameters `json:"parameters,omitempty"`
	Options    Options                  `json:"options,omitempty"`
	Model      string                   `json:"-"`
}

type ImageText2TextResponse []struct {
	GeneratedText string `json:"generated_text,omitempty"`
}

// Image-text-to-text models take in an image and text prompt and output text.
// These models are also called vision-language models, or VLMs.
// The difference from image-to-text models is that these models take an additional text input, not restricting the model to certain use cases like image captioning, and may also be trained to accept a conversation as input.

// ImageText2Text performs image-text-to-text using the specified model.
// It sends a POST request to the Hugging Face inference endpoint with the provided inputs.
// The response contains the generated text or an error if the request fails.
func (ic *InferenceClient) ImageText2Text(ctx context.Context, req *ImageText2TextRequest) (ImageText2TextResponse, error) {
	if req.Inputs == "" {
		return nil, errors.New("inputs are required")
	}

	body, err := ic.post(ctx, req.Model, "text2text-generation", req)
	if err != nil {
		return nil, err
	}

	Imagetext2TextResponse := ImageText2TextResponse{}
	if err := json.Unmarshal(body, &Imagetext2TextResponse); err != nil {
		return nil, err
	}

	return Imagetext2TextResponse, nil
}

// MODEL naver-clova-ix/donut-base-finetuned-docvqa

// https://github.com/huggingface/huggingface_hub/blob/1ae2337d26a683ee4e63c9c0b0a9f158b70d7220/tests/cassettes/TestInferenceClient.test_document_question_answering%5Bhf-inference%2Cdocument-question-answering%5D.yaml#L17

// body: '{
// "inputs": {
// 		"question": "What is the purchase amount?",
// 		"image": "base64 image..."
// },
// "parameters": {}
// }'
