namespace QuakeStats.Domain.Events;

public record PlayerDisconnectEvent : BaseEvent
{
    public override string EventType => "PLAYER_DISCONNECT";

    public required string Name { get; init; }
    public required string StreamId { get; init; }
}
