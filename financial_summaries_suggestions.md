# Administrative Financial Summaries & Reporting Suggestions

Based on the cooperative's data model (`Members`, `Dues`, `Fines`, `Events`, `Contributions`, and `Attendance`), here are suggestions for powerful financial summaries that would provide great value to the admin, as well as recommended PDF statements.

## 1. Top-Level Admin Dashboard Summaries

These are at-a-glance metrics that give the admin an immediate understanding of the committee's financial health.

- **Total Treasury Balance:** The sum of all collected inflows (`paid` dues + `paid` fines + `paid` event contributions). 
- **Total Outstanding Receivables (Arrears):** The sum of all expected money that is currently missing (`pending`/`owed`/`partially_paid` dues + `outstanding` fines + `pending` event contributions).
- **Collection Efficiency Rate:** The percentage of total expected funds (Total Treasury Balance / (Total Treasury Balance + Total Outstanding Receivables)) that have been successfully collected. This tracks how compliant the committee is.
- **Event Funding Progress:** For active events, a progress bar showing `Total Collected` vs the `goal_amount` (from `min_amount_expected` per member).
- **Recent Inflows (Last 30 Days):** Total amount collected in the recent period to track current momentum.
- **At-Risk Members Count:** Number of members whose outstanding balances exceed a certain threshold (e.g., owing more than 3 months of dues or having high unapproved absence fines).

---
EXPORT ATTENDANCE FOR A SELECTED PERIOD
EXPORTCONTRIBUTIONS DATA
A button in the event dashboard called something like settled, that should mark an event as settled when all members have contributed and admin has given the money to the person being contributed for. The settled button when clicked should first bring a popup that says are you sure? with option yes and no, clicking yes is a go-ahead confirmation. Also I want a cascade of reaction when the settled button is clicked and confirmed, I want the total amount collected to be deducted from the total treasury balance but not directly as that wuld affect the current logic. in the members dashboard, I want all members who have paid for that particular event for the statuses to change to settled when the setted button is click and confirmed in the event dashboard.

## 2. Recommended PDF Statements

PDF generation is a great way to maintain official records and send automated updates. Here are four types of PDF statements that would be incredibly useful:

### A. The Committee Consolidated Financial Report (Periodic)
*Purpose: To give the admin and executive committee a macro-view of the organization's state.*
- **Period Summary:** (e.g., Q3 2026) Total Dues expected vs collected, Fines levied vs paid, Contributions expected vs paid.
- **Asset Overview:** Total historical collections across all categories.
- **Event Highlights:** Summary of funding for ongoing or recently concluded events.
- **Compliance Overview:** The percentage of "Bonafide" members in good standing vs those in arrears.

### B. Comprehensive Member Statement of Account (Individual)
*Purpose: To be sent to individual members so they understand their exact financial standing and obligations.*
- **Header:** Member Name, Status, Join Date, and "Good Standing / Bonafide" Status.
- **Account Summary:** 
  - Total Paid to Date
  - Current Outstanding Balance (The total amount they owe)
- **Itemized Ledgers:**
  - **Dues Ledger:** Breakdown of expected dues, dates, and payment status (Paid, Partially Paid, Pending).
  - **Fines Ledger:** List of fines levied (e.g., "Unapproved absence on [Date]"), amounts, and status.
  - **Event Contributions:** List of events, expected minimum contribution, amount paid, and status.
- **Footer/Notes:** Reminders on how to pay or whom to contact to clear outstanding balances.

### C. The Arrears & Debt Aging Report (Admin Only)
*Purpose: To help the admin actively chase down pending funds.*
- **Master Debt List:** A sorted table of members with outstanding balances, ordered from highest debtor to lowest.
- **Category Breakdown:** For each debtor, how much of the debt is from Dues vs Fines vs Event Contributions.
- **Debt Aging:** An indicator of how long the oldest debt has been pending (e.g., 30+ days, 60+ days), which is especially useful for unpaid dues and older fines.

### D. Event Final Audit Report (Per Event)
*Purpose: To review the success of specific financial goals or projects.*
- **Header:** Event Title, Date, Description, and Goal Amount.
- **Funding Summary:** Total amount collected vs Goal Amount (Deficit/Surplus).
- **Contribution Roll Call:** A clear list of all members, showing who paid their expected contribution, who over-contributed, and who defaulted.
