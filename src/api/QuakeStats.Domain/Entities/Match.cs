using QuakeStats.Domain.Enums;
using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Entities;

public class Match : BaseEntity
{
    public MatchState State { get; set; }
    public Guid MatchGuid { get; set; }
    public required string Map { get; set; }
    public GameType GameType { get; set; }

    public void Apply(MatchStartedEvent @event)
    {
        MatchGuid = @event.MatchGuid;
        Map = @event.Map;
        GameType = @event.GameType;
        State = MatchState.Started;
    }

    public void Apply(MatchReportEvent @event)
    {
        MatchGuid = @event.MatchGuid;
        Map = @event.Map;
        GameType = @event.GameType;
        State = @event.Aborterd ? MatchState.Aborted : MatchState.Completed;
    }
}
