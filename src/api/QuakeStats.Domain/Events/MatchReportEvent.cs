using QuakeStats.Domain.Enums;

namespace QuakeStats.Domain.Events;

public record MatchReportEvent : BaseEvent
{
    public override string EventType => "MATCH_STARTED";
    public bool Aborterd { get; set; }
    public GameType GameType { get; set; }
    public required string Map { get; set; }
}
