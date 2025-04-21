using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using QuakeStats.Domain.Services;
using QuakeStats.Infrastructure.Data;

namespace QuakeStats.Infrastructure.Services;

public class EventProcessingService(
    IServiceProvider serviceProvider,
    ILogger<EventProcessingService> logger) : BackgroundService
{
    private readonly TimeSpan pollingInterval = TimeSpan.FromSeconds(10);

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        logger.LogInformation("Event Processing Service is starting");

        while (!stoppingToken.IsCancellationRequested)
        {
            try
            {
                await ProcessPendingEventsAsync(stoppingToken);
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Error occurred while processing events");
            }

            await Task.Delay(pollingInterval, stoppingToken);
        }

        logger.LogInformation("Event Processing Service is stopping");
    }

    private async Task ProcessPendingEventsAsync(CancellationToken stoppingToken)
    {
        using var scope = serviceProvider.CreateScope();
        var dbContext = scope.ServiceProvider.GetRequiredService<ApplicationDbContext>();
        var eventProcessor = scope.ServiceProvider.GetRequiredService<IEventProcessor>();

        var pendingEvents = await dbContext.Events
            .Where(e => !e.Processed)
            .OrderBy(e => e.CreatedAt)
            .Take(10) // Process in batches to avoid large transactions
            .ToListAsync(stoppingToken);

        if (pendingEvents.Count == 0)
        {
            return;
        }

        logger.LogInformation("Found {Count} events to process", pendingEvents.Count);

        foreach (var @event in pendingEvents)
        {
            try
            {
                await eventProcessor.ProcessEventAsync(@event);
                dbContext.Update(@event);
            }
            catch (Exception ex)
            {
                logger.LogError(ex, "Failed to process event {EventId}", @event.Id);
            }
        }

        await dbContext.SaveChangesAsync(stoppingToken);
    }
}
