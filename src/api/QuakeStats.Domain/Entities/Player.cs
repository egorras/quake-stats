using QuakeStats.Domain.Events;

namespace QuakeStats.Domain.Entities;

public class Player : BaseEntity
{
    public required string Name { get; set; }
    public required string SteamId { get; set; }

    public void Apply(PlayerConnectEvent @event)
    {
        Name = @event.Name;
        SteamId = @event.SteamId;
    }

    public void Apply(PlayerDisconnectEvent @event)
    {
        Name = @event.Name;
        SteamId = @event.SteamId;
    }
}
