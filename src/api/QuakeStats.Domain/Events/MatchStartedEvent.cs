using QuakeStats.Domain.Enums;

namespace QuakeStats.Domain.Events;

public record MatchStartedEvent : BaseEvent
{
    public GameType GameType { get; set; }
    public required string Map { get; init; }
}
