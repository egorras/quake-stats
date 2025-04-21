using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Entities;
using QuakeStats.Domain.Services;

namespace QuakeStats.Infrastructure.Services;

public class EventProcessor(
    IEnumerable<IEventHandler> handlers,
    ILogger<EventProcessor> logger) : IEventProcessor
{
    public async Task ProcessEventAsync(Event @event)
    {
        try
        {
            logger.LogInformation("Processing event {EventId} of type {EventType}", @event.Id, @event.EventType);

            // Find a handler that can handle this event type
            var handler = handlers.FirstOrDefault(h => h.CanHandle(@event.EventType));
            if (handler == null)
            {
                logger.LogWarning("No handler found for event type: {EventType}", @event.EventType);
                @event.Processed = true; // Mark as processed even if no handler found
                return;
            }

            // Parse the event data and handle it
            await handler.HandleAsync(@event);

            // Mark as processed
            @event.Processed = true;

            logger.LogInformation("Successfully processed event {EventId}", @event.Id);
        }
        catch (Exception ex)
        {
            logger.LogError(ex, "Error processing event {EventId}", @event.Id);
            throw;
        }
    }
}
