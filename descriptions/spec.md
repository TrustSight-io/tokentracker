Task: Develop a Golang Token Usage Tracking Module for LLM Calls
Objective

Create a reusable Golang module/interface that can accurately track token usage and calculate pricing for API calls to various LLM providers (Gemini, Claude, OpenAI) within our agent-media platform.
Requirements
Core Functionality

    Create a token counting service that supports all our LLM providers:
        Gemini (primary)
        Claude (fallback)
        OpenAI (fallback)

    Implement token counting methods for:
        Text input tokens
        Messages in chat format
        Response tokens
        Tool calls/function calling

    Provide price calculation based on model-specific pricing

Interface Design

// Main interface for token counting and pricing
type TokenTracker interface {
    // Count tokens for a text string or chat messages
    CountTokens(params TokenCountParams) (TokenCount, error)
    
    // Calculate price based on token usage
    CalculatePrice(model string, inputTokens int, outputTokens int) (Price, error)
    
    // Track full usage for an LLM call (input and output)
    TrackUsage(callParams CallParams, response interface{}) (UsageMetrics, error)
}

// Supporting types
type TokenCountParams struct {
    Model            string
    Text             *string
    Messages         []Message
    Tools            []Tool
    ToolChoice       *ToolChoice
    CountResponseTokens bool
}

type TokenCount struct {
    InputTokens      int
    ResponseTokens   int
    TotalTokens      int
}

type Price struct {
    InputCost        float64
    OutputCost       float64
    TotalCost        float64
    Currency         string
}

type UsageMetrics struct {
    TokenCount       TokenCount
    Price            Price
    Duration         time.Duration
    Timestamp        time.Time
    Model            string
    Provider         string
}

Provider-Specific Implementation

    Gemini Implementation
        Integrate with Gemini's token counting mechanism
        Support all Gemini models we use

    Claude Implementation
        Implement Anthropic's token counting logic
        Support Claude 3 Haiku, Sonnet, and Opus models

    OpenAI Implementation
        Leverage tiktoken or equivalent for accurate OpenAI token counting
        Support GPT-3.5 and GPT-4 models

Configuration System

    Create a flexible configuration system that allows:
        Setting per-model pricing
        Updating pricing without code changes
        Customizing default values

Error Handling and Logging

    Implement robust error handling for all token counting scenarios
    Log token usage and pricing information for auditing purposes
    Provide fallback mechanisms for when token counting fails

Performance Considerations

    Minimize latency impact in the request flow
    Implement caching where appropriate for repeated token calculations
    Support concurrent usage across multiple microservices

Integration with Existing Infrastructure

    Design the module to work seamlessly with our Go microservices architecture
    Provide hooks for sending usage data to our monitoring/analytics systems
    Support both synchronous and asynchronous operation modes

Deliverables

    Complete Go module/package with all required functionality
    Unit and integration tests covering all primary code paths
    Example implementation for each supported LLM provider
    Documentation and usage examples
    Performance benchmarks

Technical Constraints

    The module should be thread-safe
    No external HTTP calls for token calculation (except when absolutely necessary)
    Minimal dependencies to reduce maintenance overhead
    Must work with Go 1.21+

Additional Notes

    The implementation should be inspired by LiteLLM's approach but optimized for Go : (https://github.com/BerriAI/litellm/blob/92883560f03c8044474165fa15b36f0564a1d570/litellm/utils.py#L1837)
    There is a function definition on the highlighted line number :
    """
def token_counter(
    model="",
    custom_tokenizer: Optional[Union[dict, SelectTokenizerResponse]] = None,
    text: Optional[Union[str, List[str]]] = None,
    messages: Optional[List] = None,
    count_response_tokens: Optional[bool] = False,
    tools: Optional[List[ChatCompletionToolParam]] = None,
    tool_choice: Optional[ChatCompletionNamedToolChoiceParam] = None,
    use_default_image_token_count: Optional[bool] = False,
    default_token_count: Optional[int] = None,
) -> int:
    """
    Count the number of tokens in a given text using a specified model.

    Args:
    model (str): The name of the model to use for tokenization. Default is an empty string.
    custom_tokenizer (Optional[dict]): A custom tokenizer created with the `create_pretrained_tokenizer` or `create_tokenizer` method. Must be a dictionary with a string value for `type` and Tokenizer for `tokenizer`. Default is None.
    text (str): The raw text string to be passed to the model. Default is None.
    messages (Optional[List[Dict[str, str]]]): Alternative to passing in text. A list of dictionaries representing messages with "role" and "content" keys. Default is None.
    default_token_count (Optional[int]): The default number of tokens to return for a message block, if an error occurs. Default is None.

    Returns:
    int: The number of tokens in the text.
    """
    # use tiktoken, anthropic, cohere, llama2, or llama3's tokenizer depending on the model
    is_tool_call = False
    num_tokens = 0
    if text is None:
        if messages is not None:
            print_verbose(f"token_counter messages received: {messages}")
            text = ""
            for message in messages:
                if message.get("content", None) is not None:
                    content = message.get("content")
                    if isinstance(content, str):
                        text += message["content"]
                    elif isinstance(content, List):
                        text, num_tokens = _get_num_tokens_from_content_list(
                            content_list=content,
                            use_default_image_token_count=use_default_image_token_count,
                            default_token_count=default_token_count,
                        )
                if message.get("tool_calls"):
                    is_tool_call = True
                    for tool_call in message["tool_calls"]:
                        if "function" in tool_call:
                            function_arguments = tool_call["function"]["arguments"]
                            text = (
                                text if isinstance(text, str) else "".join(text or [])
                            ) + (str(function_arguments) if function_arguments else "")

        else:
            raise ValueError("text and messages cannot both be None")
    elif isinstance(text, List):
        text = "".join(t for t in text if isinstance(t, str))
    elif isinstance(text, str):
        count_response_tokens = True  # user just trying to count tokens for a text. don't add the chat_ml +3 tokens to this

    if model is not None or custom_tokenizer is not None:
        tokenizer_json = custom_tokenizer or _select_tokenizer(model=model)
        if tokenizer_json["type"] == "huggingface_tokenizer":
            enc = tokenizer_json["tokenizer"].encode(text)
            num_tokens = len(enc.ids)
        elif tokenizer_json["type"] == "openai_tokenizer":
            if (
                model in litellm.open_ai_chat_completion_models
                or model in litellm.azure_llms
            ):
                if model in litellm.azure_llms:
                    # azure llms use gpt-35-turbo instead of gpt-3.5-turbo 🙃
                    model = model.replace("-35", "-3.5")

                print_verbose(
                    f"Token Counter - using OpenAI token counter, for model={model}"
                )
                num_tokens = openai_token_counter(
                    text=text,  # type: ignore
                    model=model,
                    messages=messages,
                    is_tool_call=is_tool_call,
                    count_response_tokens=count_response_tokens,
                    tools=tools,
                    tool_choice=tool_choice,
                    use_default_image_token_count=use_default_image_token_count
                    or False,
                    default_token_count=default_token_count,
                )
            else:
                print_verbose(
                    f"Token Counter - using generic token counter, for model={model}"
                )
                num_tokens = openai_token_counter(
                    text=text,  # type: ignore
                    model="gpt-3.5-turbo",
                    messages=messages,
                    is_tool_call=is_tool_call,
                    count_response_tokens=count_response_tokens,
                    tools=tools,
                    tool_choice=tool_choice,
                    use_default_image_token_count=use_default_image_token_count
                    or False,
                    default_token_count=default_token_count,
                )
    else:
        num_tokens = len(encoding.encode(text, disallowed_special=()))  # type: ignore
    return num_tokens

    """
    Consider using the protocol buffer format for serializing usage data
    Review Google's generative AI Go library for Gemini token counting approach
    For Claude, review Anthropic's documentation on counting tokens
