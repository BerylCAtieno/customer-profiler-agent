package profiler

import (
	"context"
	"fmt"
	"strings"

	"github.com/BerylCAtieno/customer-profiler-agent/internal/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash-lite")
	model.SetTemperature(0.7)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(2048)

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

func (g *GeminiClient) Close() {
	g.client.Close()
}

func (g *GeminiClient) GenerateCustomerProfiles(ctx context.Context, businessIdea string) (*models.ProfileResponse, error) {
	prompt := g.buildPrompt(businessIdea)

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	text := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	profile, err := g.parseSimpleProfile(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse simple profile: %w", err)
	}

	return &models.ProfileResponse{
		BusinessIdea: businessIdea,
		Profiles:     []models.CustomerProfile{*profile},
		Summary:      "",
		Keywords:     []string{},
	}, nil
}

func (g *GeminiClient) parseSimpleProfile(text string) (*models.CustomerProfile, error) {
	profile := models.CustomerProfile{}

	text = strings.TrimSpace(text)

	pairs := strings.Split(text, ", ")

	data := make(map[string]string)

	for _, pair := range pairs {
		parts := strings.SplitN(pair, ": ", 2)
		if len(parts) == 2 {
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			data[key] = value
		}
	}

	profile.Age = data["age"]
	profile.Gender = data["gender"]
	profile.Location = data["location"]
	profile.Occupation = data["occupation"]
	profile.Income = data["income"]

	profile.PainPoints = strings.Split(data["pain_points"], ",")
	profile.Motivations = strings.Split(data["motivations"], ",")
	profile.Interests = strings.Split(data["interests"], ",")

	profile.PreferredChannels = []string{data["channel"]}

	return &profile, nil
}

func (g *GeminiClient) buildPrompt(businessIdea string) string {
	return fmt.Sprintf(`You are an expert market researcher. Based ONLY on the business idea "%s", generate a SINGLE, concise customer profile.

						The output MUST be a single line of text in the format "key: value, key: value, ..." without any other text, markdown, or punctuation. Use only the following keys in this order:

						age: Age range (e.g., 30-50)
						gender: Gender (e.g., female)
						location: Geographic type (e.g., Urban)
						occupation: Job title/occupation (e.g., Marketing Manager)
						income: Income range (e.g., $75k-100k)
						pain_points: 1-2 main pain points (comma-separated, no quotes)
						motivations: 1-2 key motivations (comma-separated, no quotes)
						interests: 2-3 interests/hobbies (comma-separated, no quotes)
						channel: 1 preferred channel (e.g., Instagram)

						Example format: age: 30-50, gender: female, location: Urban, occupation: Marketing Manager, income: $75k-100k, pain_points: lack of time, overwhelming choices, motivations: convenience, quality, interests: makeup, shoes, travel, channel: Instagram`, businessIdea)
}
