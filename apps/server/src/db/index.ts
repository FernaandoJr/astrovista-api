import { createClient } from "@supabase/supabase-js"

const supabaseUrl = "https://norjgyhilqiojxvismzy.supabase.co"
const supabaseKey = process.env.ANON_KEY || ""

export const supabase = createClient(supabaseUrl, supabaseKey)
