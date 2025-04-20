using System.Text.Json.Serialization;

namespace QuakeStats.Domain.Events;

[JsonPolymorphic(TypeDiscriminatorPropertyName = "TYPE")]
[JsonDerivedEventType<PlayerConnectEvent>()]
[JsonDerivedEventType<PlayerDisconnectEvent>()]
public abstract record BaseEvent
{
    public abstract string EventType { get; }

    [JsonPropertyName("MATCH_GUID")]
    public Guid MatchId { get; init; }

    public int Time { get; init; }
    public bool Warmup { get; init; }
}
