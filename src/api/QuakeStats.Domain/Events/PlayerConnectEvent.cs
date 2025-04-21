namespace QuakeStats.Domain.Events;

public record PlayerConnectEvent : BaseEvent
{
    public required string Name { get; init; }
    public required string SteamId { get; init; }
}
