using System.Text.Json.Serialization;

namespace QuakeStats.Domain.Events;

[JsonPolymorphic(TypeDiscriminatorPropertyName = "TYPE")]
[JsonDerivedEventType<PlayerConnectEvent>()]
[JsonDerivedEventType<PlayerDisconnectEvent>()]
public abstract record BaseEvent
{
    public abstract string EventType { get; }

    public Guid MatchGuid { get; init; }
    public int Time { get; init; }
    public bool Warmup { get; init; }
}
