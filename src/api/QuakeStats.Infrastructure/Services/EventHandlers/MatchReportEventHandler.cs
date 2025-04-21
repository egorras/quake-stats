using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Events;
using QuakeStats.Infrastructure.Data;

namespace QuakeStats.Infrastructure.Services.EventHandlers;

public class MatchReportEventHandler(
    ApplicationDbContext dbContext,
    ILogger<MatchReportEventHandler> logger) : BaseEventHandler<MatchReportEvent>
{
    public override string EventType => "MATCH_REPORT";

    protected override async Task ProcessEventAsync(MatchReportEvent @event)
    {
        logger.LogInformation("Processing match report event for match {MatchGuid}", @event.MatchGuid);

        var match = await dbContext.Matches
            .FirstOrDefaultAsync(m => m.MatchGuid == @event.MatchGuid);

        if (match == null)
        {
            logger.LogWarning("Match {MatchGuid} not found for match report event", @event.MatchGuid);
            match = new();
            dbContext.Matches.Add(match);
        }

        match.Apply(@event);
    }
}
