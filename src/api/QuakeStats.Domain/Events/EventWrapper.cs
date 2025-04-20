using System.Text.Json;

namespace QuakeStats.Domain.Events;

public record EventWrapper
{
    public required string TYPE { get; init; }
    public required JsonElement DATA { get; init; }


    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
        PropertyNamingPolicy = JsonNamingPolicy.KebabCaseUpper
    };

    public static EventWrapper FromJson(string json) =>
        JsonSerializer.Deserialize<EventWrapper>(json, JsonOptions)!;

    public BaseEvent GetEvent()
    {
        var discriminatorJson =
            $@"{{""TYPE"":""{TYPE}""," +
            DATA.GetRawText().TrimStart('{');

        return JsonSerializer.Deserialize<BaseEvent>(discriminatorJson, JsonOptions)
               ?? throw new JsonException("Failed to deserialize match event.");
    }
}
