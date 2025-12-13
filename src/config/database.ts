import { Pool, PoolConfig } from "pg";
import dotenv from "dotenv";

dotenv.config();

let pool: Pool | null = null;

export function initDB(): Pool {
  const databaseUrl = process.env.DATABASE_URL;

  if (!databaseUrl) {
    throw new Error("FATAL: DATABASE_URL environment variable not set.");
  }

  console.log("Connecting to database...");

  // Parse connection string
  const config: PoolConfig = {
    connectionString: databaseUrl,
    max: 25, // Maximum number of clients in the pool
    idleTimeoutMillis: 300000, // Close idle clients after 5 minutes
    connectionTimeoutMillis: 2000, // Return an error after 2 seconds if connection could not be established
  };

  pool = new Pool(config);

  // Test the connection with retry logic (async, but don't block)
  const maxRetries = 5;
  let retries = 0;

  const connectWithRetry = async (): Promise<void> => {
    try {
      const client = await pool!.connect();
      client.release();
      console.log("✅ Successfully connected to the PostgreSQL database!");
    } catch (err) {
      retries++;
      if (retries < maxRetries) {
        console.log(`Database ping attempt ${retries}/${maxRetries} failed: ${err}`);
        await new Promise((resolve) => setTimeout(resolve, 2000));
        return connectWithRetry();
      }
      console.error(`Failed to connect to database after ${maxRetries} attempts: ${err}`);
      // Don't exit here, let the app start and handle errors gracefully
    }
  };

  // Start connection check in background
  connectWithRetry();

  return pool;
}

export function getDB(): Pool {
  if (!pool) {
    throw new Error("Database not initialized. Call initDB() first.");
  }
  return pool;
}
