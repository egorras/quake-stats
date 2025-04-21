using Microsoft.EntityFrameworkCore.Migrations;

#nullable disable

namespace QuakeStats.Infrastructure.Migrations
{
    /// <inheritdoc />
    public partial class UpdateSchema : Migration
    {
        /// <inheritdoc />
        protected override void Up(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.RenameColumn(
                name: "state",
                table: "matches",
                newName: "team_score_red");

            migrationBuilder.AddColumn<DateTimeOffset>(
                name: "reported_at",
                table: "matches",
                type: "timestamp with time zone",
                nullable: true);

            migrationBuilder.AddColumn<DateTimeOffset>(
                name: "started_at",
                table: "matches",
                type: "timestamp with time zone",
                nullable: false,
                defaultValue: new DateTimeOffset(new DateTime(1, 1, 1, 0, 0, 0, 0, DateTimeKind.Unspecified), new TimeSpan(0, 0, 0, 0, 0)));

            migrationBuilder.AddColumn<int>(
                name: "team_score_blue",
                table: "matches",
                type: "integer",
                nullable: false,
                defaultValue: 0);
        }

        /// <inheritdoc />
        protected override void Down(MigrationBuilder migrationBuilder)
        {
            migrationBuilder.DropColumn(
                name: "reported_at",
                table: "matches");

            migrationBuilder.DropColumn(
                name: "started_at",
                table: "matches");

            migrationBuilder.DropColumn(
                name: "team_score_blue",
                table: "matches");

            migrationBuilder.RenameColumn(
                name: "team_score_red",
                table: "matches",
                newName: "state");
        }
    }
}
