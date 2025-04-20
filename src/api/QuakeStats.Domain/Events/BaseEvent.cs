using System.Text.Json.Serialization;
using QuakeStats.Domain.Events.Attributes;

namespace QuakeStats.Domain.Events;

[JsonPolymorphic(TypeDiscriminatorPropertyName = "TYPE")]
[JsonDerivedEventType<MatchStartedEvent>()]
[JsonDerivedEventType<PlayerConnectEvent>()]
[JsonDerivedEventType<PlayerDisconnectEvent>()]
public abstract record BaseEvent
{
    public abstract string EventType { get; }

    public Guid MatchGuid { get; init; }
    public int Time { get; init; }
    public bool Warmup { get; init; }
}
