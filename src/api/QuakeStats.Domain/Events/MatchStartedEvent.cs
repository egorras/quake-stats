namespace QuakeStats.Domain.Events;

public record MatchStartedEvent : BaseEvent
{
    public override string EventType => "MATCH_STARTED";
    public required string Map { get; init; }
}
