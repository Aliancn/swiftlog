# OpenAI API Configuration Guide

This guide explains how to configure OpenAI API for SwiftLog's AI analysis features.

## Overview

OpenAI API configuration is **user-level**, not system-level. Each user configures their own API key through the web interface. The system does not require any OpenAI configuration to start - it's completely optional.

**Key Points:**
- ✅ Configuration is per-user, stored in the database
- ✅ System starts without any OpenAI configuration
- ✅ Users configure through the Settings page in the web UI
- ✅ No need to edit .env files or restart services
- ✅ Supports project-level overrides

## Configuration Methods

### Method 1: Web UI (Recommended)

**Steps:**

1. **Login to SwiftLog**
   - Open your SwiftLog instance in a web browser
   - Login with your credentials

2. **Navigate to Settings**
   - Click on your user menu in the top-right corner
   - Select "Settings"

3. **Configure AI Settings**
   - In the "AI Configuration" section, fill in:
     - **AI API Key**: Your OpenAI API key (e.g., `sk-...`)
     - **AI Base URL**: API endpoint (default: `https://api.openai.com/v1`)
     - **AI Model**: Model to use (default: `gpt-4o-mini`)
     - **Max Tokens**: Maximum tokens for analysis (default: 4000)
     - **Prompt Language**: Language for AI prompts (English/Chinese)
     - **AI Enabled**: Toggle to enable/disable AI features
     - **Auto Analyze**: Automatically analyze runs when they complete

4. **Save Configuration**
   - Click "Save Settings"
   - Configuration takes effect immediately - no restart needed

5. **Test the Configuration**
   - Navigate to any run page
   - Click "AI Analysis" button
   - Verify that analysis generates successfully

**Benefits of Web UI Configuration:**
- ✅ Each user has their own API key and settings
- ✅ No need to access the server or edit files
- ✅ Changes take effect immediately
- ✅ Secure - keys stored encrypted in database
- ✅ Project-level overrides supported

### Method 2: API Configuration (Advanced)

For programmatic configuration, you can use the API directly:

```bash
# Get current settings
curl -X GET https://your-swiftlog.com/api/v1/settings \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Update settings
curl -X PUT https://your-swiftlog.com/api/v1/settings \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_enabled": true,
    "ai_api_key": "sk-your-key",
    "ai_base_url": "https://api.openai.com/v1",
    "ai_model": "gpt-4o-mini"
  }'
```

## Configuration Settings

These settings are configured per-user through the web UI:

| Setting | Required | Default | Description |
|----------|----------|---------|-------------|
| `AI API Key` | Yes* | - | Your OpenAI API key |
| `AI Base URL` | No | `https://api.openai.com/v1` | API endpoint (for Azure OpenAI, custom endpoints) |
| `AI Model` | No | `gpt-4o-mini` | Model to use for analysis |
| `Max Tokens` | No | `4000` | Maximum tokens for AI responses |
| `Prompt Language` | No | `English` | Language for AI prompts (English/Chinese) |
| `AI Enabled` | No | `true` | Enable/disable AI features for this user |
| `Auto Analyze` | No | `false` | Automatically analyze runs on completion |

*Required only if you want to enable AI analysis features

## Getting an OpenAI API Key

1. **Sign up for OpenAI**
   - Go to https://platform.openai.com/
   - Create an account or sign in

2. **Create an API Key**
   - Navigate to API Keys section
   - Click "Create new secret key"
   - Copy the key (you won't be able to see it again)

3. **Add Credits (if needed)**
   - Go to Billing section
   - Add payment method
   - Purchase credits

## Alternative API Providers

### Azure OpenAI

Configure in the Settings page with:
- **AI API Key**: Your Azure OpenAI API key
- **AI Base URL**: `https://your-resource.openai.azure.com/`
- **AI Model**: `gpt-4` (or your deployed model name)

### OpenAI-Compatible APIs (LocalAI, Ollama, etc.)

Configure in the Settings page with:
- **AI API Key**: `dummy-key` (if authentication not required)
- **AI Base URL**: `http://localhost:8080/v1` (your local API endpoint)
- **AI Model**: `llama2` (or your model name)

**Note**: These alternatives work the same way - just configure different URLs and models in your user settings.

## Verification

After configuration, verify that AI features are working:

### Check Configuration in Web UI

1. **Go to Settings Page**
   - You should see a green indicator "● Key configured" next to the API Key field
   - This means your API key is stored in the database

2. **Verify Settings are Saved**
   - Refresh the Settings page
   - Check that your configuration is still there
   - API key will be masked as `••••••••••••••••`

### Test AI Analysis

1. **Navigate to a Run Page**
   - Go to any project with completed runs
   - Open a run detail page

2. **Trigger AI Analysis**
   - Click "AI Analysis" or "Generate Report" button
   - Wait for the analysis to complete

3. **Check Results**
   - If configured correctly, you'll see:
     - AI analysis summary
     - Key findings
     - Recommendations
   - If there's an error, check the error message for details

### Troubleshoot if Needed

If AI analysis doesn't work:

1. **Check Settings**
   - Go back to Settings page
   - Verify all fields are filled correctly
   - Make sure "AI Enabled" is toggled ON

2. **Check API Key**
   - Verify your API key is valid
   - Check if you have credits in your OpenAI account
   - Test the key directly via OpenAI API if needed

3. **Check System Logs** (for administrators)
   ```bash
   # View AI worker logs
   docker compose logs ai-worker

   # Look for errors
   docker compose logs ai-worker | grep -i error
   ```

## Troubleshooting

### Error: "OpenAI API key not configured"

**Cause:** You haven't set up your API key yet, or it wasn't saved properly.

**Solution:**
1. Go to Settings page
2. Enter your OpenAI API key in the "AI API Key" field
3. Make sure to click "Save Settings"
4. Refresh the page to verify the green "● Key configured" indicator appears

### Error: "Invalid API key"

**Causes:**
- API key is incorrect or has typos
- API key has been revoked in OpenAI dashboard
- Extra spaces or characters in the key

**Solution:**
1. Go to OpenAI dashboard and verify your API key
2. If needed, generate a new API key
3. In SwiftLog Settings page:
   - Clear the existing API key field
   - Paste the new key (make sure no extra spaces)
   - Save settings
4. Test again

### Error: "Rate limit exceeded"

**Causes:**
- Too many requests
- Free tier limit reached
- Insufficient credits

**Solution:**
1. Check your OpenAI usage dashboard
2. Add credits if needed
3. Reduce analysis frequency
4. Consider using a different model

### AI Features Not Working

**Step-by-step diagnosis:**

1. **Check Your Settings**
   - Go to Settings page
   - Verify "AI Enabled" is ON
   - Verify API key is configured (green indicator)
   - Check that all required fields are filled

2. **Test API Key Separately**
   - Try using your API key in a simple curl test:
   ```bash
   curl https://api.openai.com/v1/chat/completions \
     -H "Authorization: Bearer sk-your-key" \
     -H "Content-Type: application/json" \
     -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hi"}]}'
   ```
   - If this fails, the issue is with your API key/account

3. **Check AI Worker Status** (for administrators)
   ```bash
   docker compose ps ai-worker
   docker compose logs ai-worker
   ```

**Common issues:**
- AI worker not running → `docker compose up -d ai-worker`
- Network issues → Check server internet connectivity
- Database connection issues → Check database is running

## Security Best Practices

### 1. Protect Your API Key

- ✅ API keys are stored encrypted in the database
- ✅ Keys are never exposed in the UI (shown as `••••••••`)
- ✅ Keys are never logged by the application
- ⚠️ Never share your API key with others
- ⚠️ Don't commit API keys to version control

### 2. Rotate Keys Regularly

**Recommended rotation schedule:** Every 90 days

**How to rotate:**
1. Generate a new key in OpenAI dashboard
2. In SwiftLog Settings page:
   - Update your API key with the new one
   - Save settings
3. Test that AI features still work
4. Revoke the old key in OpenAI dashboard

### 3. Monitor Usage

- Set up usage alerts in OpenAI dashboard
- Monitor monthly spending
- Review API usage logs in OpenAI platform
- Check SwiftLog AI analysis history

### 4. Use Different Keys for Different Environments

**Best practice for organizations:**
- Production users: Use production OpenAI account
- Staging/testing users: Use separate testing account
- Development: Use development account with lower limits

**Advantages:**
- Separate billing and cost tracking
- Isolate production from testing
- Apply different rate limits per environment

## Cost Management

### Estimate Costs

- Model: gpt-4o-mini
- Approximate cost per analysis: $0.01 - $0.05
- Monitor via OpenAI dashboard

### Reduce Costs

1. **Use cheaper models:**
   ```bash
   OPENAI_MODEL=gpt-3.5-turbo  # Cheaper than gpt-4
   ```

2. **Limit analysis frequency**
   - Only analyze failed runs
   - Manual triggers only

3. **Set usage limits**
   - Configure limits in OpenAI dashboard
   - Monitor and adjust

## Disabling AI Features

### For Individual Users

To disable AI features for yourself:

1. Go to Settings page
2. Toggle "AI Enabled" to OFF
3. Save settings
4. AI features will be disabled for your account only

### For Entire System (Administrators)

To disable AI for the entire system:

```bash
# Stop AI worker
docker compose stop ai-worker

# Prevent it from starting on boot
# Edit docker-compose.yaml and comment out ai-worker service
```

**Note:** Even with the AI worker stopped, users can still save their API keys in settings - they just won't be used until the AI worker is restarted.

## Advanced Configuration

### Project-Level Settings Override

Administrators can set project-specific AI configurations that override user settings:

**Via API:**
```bash
curl -X PUT https://your-swiftlog.com/api/v1/projects/{project_id}/settings \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "ai_model": "gpt-4",
    "ai_max_tokens": 8000
  }'
```

**Use case:** Force all analyses for a specific project to use a more powerful model.

### Custom Prompts

AI prompts can be customized in the backend code. See `backend/internal/ai/analyzer.go` for implementation details.

### Multiple Models

Each user can configure their own preferred model:
- Users can choose different models based on their needs
- Common options: `gpt-4o-mini` (cheaper), `gpt-4o` (more capable)
- Just change the "AI Model" setting in the Settings page

## Support

For issues related to:
- **OpenAI API**: Check OpenAI documentation or support
- **SwiftLog configuration**: Check main documentation
- **Integration issues**: Review AI worker logs

## References

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [OpenAI Pricing](https://openai.com/pricing)
- [SwiftLog Environment Variables](ENVIRONMENT_VARIABLES.md)
- [SwiftLog Deployment Guide](DEPLOYMENT.md)
