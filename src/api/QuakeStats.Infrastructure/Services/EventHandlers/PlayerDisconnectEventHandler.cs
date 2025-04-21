using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Events;
using QuakeStats.Infrastructure.Data;

namespace QuakeStats.Infrastructure.Services.EventHandlers;

public class PlayerDisconnectEventHandler(
    ApplicationDbContext dbContext,
    ILogger<PlayerDisconnectEventHandler> logger) : BaseEventHandler<PlayerDisconnectEvent>
{
    public override string EventType => "PLAYER_DISCONNECT";

    protected override async Task ProcessEventAsync(PlayerDisconnectEvent @event)
    {
        logger.LogInformation("Processing player disconnect event for player {SteamId}", @event.SteamId);

        var player = await dbContext.Players
            .FirstOrDefaultAsync(p => p.SteamId == @event.SteamId);

        if (player == null)
        {
            player = new();
            dbContext.Players.Add(player);
            logger.LogWarning("Player {SteamId} not found for disconnect event", @event.SteamId);
            return;
        }

        player.Apply(@event);
    }
}
