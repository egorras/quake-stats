namespace QuakeStats.Domain.Events;

public record PlayerConnectEvent : BaseEvent
{
    public override string EventType => "PLAYER_CONNECT";

    public required string Name { get; init; }
    public required string StreamId { get; init; }
}
