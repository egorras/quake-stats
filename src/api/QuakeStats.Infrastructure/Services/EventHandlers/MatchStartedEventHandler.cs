using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Events;
using QuakeStats.Infrastructure.Data;

namespace QuakeStats.Infrastructure.Services.EventHandlers;

public class MatchStartedEventHandler(
    ApplicationDbContext dbContext,
    ILogger<MatchStartedEventHandler> logger) : BaseEventHandler<MatchStartedEvent>
{
    public override string EventType => "MATCH_STARTED";

    protected override async Task ProcessEventAsync(MatchStartedEvent @event)
    {
        logger.LogInformation("Processing match started event for match {MatchGuid}", @event.MatchGuid);

        var match = await dbContext.Matches
            .FirstOrDefaultAsync(m => m.MatchGuid == @event.MatchGuid);

        if (match == null)
        {
            match = new();
            dbContext.Matches.Add(match);
        }

        match.Apply(@event);
    }
}
