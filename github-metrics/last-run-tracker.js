import { readFile, writeFile, mkdir } from 'fs/promises';
import { existsSync } from 'fs';
import { join, dirname } from 'path';

const DATA_DIR = '/data';
const LAST_RUN_FILE = join(DATA_DIR, 'last-run.json');
const FOURTEEN_DAYS_MS = 14 * 24 * 60 * 60 * 1000; // 14 days in milliseconds

/**
 * Checks if enough time has passed since the last run.
 * @returns {Promise<{shouldRun: boolean, lastRun: Date|null, daysSinceLastRun: number|null}>}
 */
export async function shouldRunMetricsCollection() {
    try {
        // Ensure data directory exists
        if (!existsSync(DATA_DIR)) {
            await mkdir(DATA_DIR, { recursive: true });
        }

        // Check if last run file exists
        if (!existsSync(LAST_RUN_FILE)) {
            console.log('No previous run detected. This is the first run.');
            return {
                shouldRun: true,
                lastRun: null,
                daysSinceLastRun: null
            };
        }

        // Read last run timestamp
        const data = await readFile(LAST_RUN_FILE, 'utf8');
        const { lastRun } = JSON.parse(data);
        const lastRunDate = new Date(lastRun);
        const now = new Date();
        const timeSinceLastRun = now - lastRunDate;
        const daysSinceLastRun = Math.floor(timeSinceLastRun / (24 * 60 * 60 * 1000));

        console.log(`Last run: ${lastRunDate.toISOString()}`);
        console.log(`Days since last run: ${daysSinceLastRun}`);

        if (timeSinceLastRun >= FOURTEEN_DAYS_MS) {
            console.log(`✓ 14 days have passed. Proceeding with metrics collection.`);
            return {
                shouldRun: true,
                lastRun: lastRunDate,
                daysSinceLastRun
            };
        } else {
            const daysRemaining = Math.ceil((FOURTEEN_DAYS_MS - timeSinceLastRun) / (24 * 60 * 60 * 1000));
            console.log(`✗ Only ${daysSinceLastRun} days have passed. Skipping metrics collection.`);
            console.log(`  Next run should occur in approximately ${daysRemaining} days.`);
            return {
                shouldRun: false,
                lastRun: lastRunDate,
                daysSinceLastRun
            };
        }
    } catch (error) {
        console.error('Error checking last run time:', error);
        // If there's an error reading the file, assume we should run
        console.log('Error reading last run file. Proceeding with metrics collection.');
        return {
            shouldRun: true,
            lastRun: null,
            daysSinceLastRun: null
        };
    }
}

/**
 * Records the current timestamp as the last successful run.
 * @returns {Promise<void>}
 */
export async function recordSuccessfulRun() {
    try {
        // Ensure data directory exists
        if (!existsSync(DATA_DIR)) {
            await mkdir(DATA_DIR, { recursive: true });
        }

        const now = new Date();
        const data = {
            lastRun: now.toISOString(),
            timestamp: now.getTime()
        };

        await writeFile(LAST_RUN_FILE, JSON.stringify(data, null, 2), 'utf8');
        console.log(`✓ Recorded successful run at ${now.toISOString()}`);
    } catch (error) {
        console.error('Error recording last run time:', error);
        // Don't throw - we don't want to fail the entire job just because we couldn't write the file
    }
}

/**
 * Gets information about the last run without checking if we should run.
 * Useful for debugging and monitoring.
 * @returns {Promise<{lastRun: Date|null, daysSinceLastRun: number|null}>}
 */
export async function getLastRunInfo() {
    try {
        if (!existsSync(LAST_RUN_FILE)) {
            return {
                lastRun: null,
                daysSinceLastRun: null
            };
        }

        const data = await readFile(LAST_RUN_FILE, 'utf8');
        const { lastRun } = JSON.parse(data);
        const lastRunDate = new Date(lastRun);
        const now = new Date();
        const timeSinceLastRun = now - lastRunDate;
        const daysSinceLastRun = Math.floor(timeSinceLastRun / (24 * 60 * 60 * 1000));

        return {
            lastRun: lastRunDate,
            daysSinceLastRun
        };
    } catch (error) {
        console.error('Error getting last run info:', error);
        return {
            lastRun: null,
            daysSinceLastRun: null
        };
    }
}

