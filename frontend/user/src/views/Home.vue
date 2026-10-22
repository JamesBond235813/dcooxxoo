<template>
  <div class="home-page min-h-screen bg-background text-foreground">

    <!-- ==================== HERO SECTION ==================== -->
    <section class="relative z-10 pt-20 pb-3 md:pt-24 md:pb-8">
      <div class="max-w-6xl mx-auto px-4">
        <div class="mb-3 flex justify-end"><PlusHalfCheck /></div>
        <div class="relative isolate min-h-[150px] overflow-hidden rounded-[1.35rem] border border-emerald-200/70 bg-[radial-gradient(ellipse_at_top_right,_rgba(255,255,255,0.96),_transparent_58%),linear-gradient(135deg,#f0f8f2_0%,#e8f2ed_56%,#edf5f8_100%)] px-4 py-4 shadow-[0_10px_28px_-22px_rgba(16,85,65,0.45)] dark:border-emerald-900/50 dark:bg-[linear-gradient(135deg,#14251d,#182a26_58%,#17242b)] md:hidden">
          <div class="absolute -right-8 -top-10 -z-10 h-32 w-32 rounded-full bg-emerald-200/30 blur-2xl dark:bg-emerald-500/10"></div>
          <div class="flex items-center justify-between gap-3">
            <span class="inline-flex items-center gap-1.5 rounded-full border border-emerald-200/80 bg-white/75 px-2.5 py-1 text-[10px] font-semibold tracking-wide text-emerald-800 shadow-xs dark:border-emerald-800 dark:bg-emerald-950/50 dark:text-emerald-300"><span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>{{ t('homeMobile.heroTag') }}</span>
            <span class="text-right text-[10px] font-semibold tracking-wide text-emerald-800 dark:text-emerald-300">新时代水电煤<br class="hidden max-[370px]:block" /> 有它就有超能力</span>
          </div>
          <div class="mt-4 flex items-end justify-between gap-3">
            <div class="min-w-0">
              <h1 class="text-[22px] font-extrabold tracking-[-0.035em] text-slate-900 dark:text-white">GPT 源头代充</h1>
              <p class="mt-1 text-xs font-medium tracking-wide text-slate-600 dark:text-slate-300">{{ t('homeMobile.heroSubtitle') }}</p>
            </div>
            <div class="flex shrink-0 items-center gap-3 pb-0.5" aria-label="OpenAI 与 Claude">
              <img src="/openai.svg" alt="OpenAI" class="h-8 w-8 object-contain dark:invert" />
              <span class="h-7 w-px bg-slate-300/80 dark:bg-slate-600"></span>
              <img src="/claude.svg" alt="Claude（即将推出）" class="h-8 w-8 object-contain grayscale opacity-45" />
            </div>
          </div>
        </div>
        <div class="relative hidden overflow-hidden rounded-3xl border border-emerald-200/60 dark:border-emerald-900/40 bg-[#ecf3ec] dark:bg-[#15231c] p-5 sm:p-6 md:block md:p-7 shadow-sm">
          <div class="flex flex-col md:flex-row items-center justify-between gap-5 md:gap-8">
            <!-- Left Info -->
            <div class="space-y-2 md:space-y-3 text-center md:text-left flex-1">
              <div class="inline-flex items-center gap-2 rounded-full border border-emerald-300/60 dark:border-emerald-800 bg-emerald-100/90 dark:bg-emerald-950/80 px-3 py-1 text-xs font-semibold text-emerald-800 dark:text-emerald-300">
                <span class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></span>
                <span>新时代的水电煤 学习工作效率生产力拉满</span>
              </div>

              <h1 class="text-3xl sm:text-4xl md:text-[2.65rem] font-extrabold tracking-tight text-gray-900 dark:text-white">
                GPT 源头代充
              </h1>

              <p class="text-sm sm:text-base md:text-lg font-medium text-gray-600 dark:text-gray-300">
                稳定 · 高效 · 长期服务
              </p>

              <!-- Feature Tags -->
              <div class="flex flex-wrap items-center justify-center md:justify-start gap-2.5 pt-1">
                <span class="inline-flex items-center gap-1.5 rounded-full border border-gray-300/60 dark:border-gray-700 bg-white/80 dark:bg-black/20 px-3.5 py-1 text-xs font-medium text-gray-700 dark:text-gray-200 backdrop-blur-xs">
                  <span class="text-emerald-600">✓</span> 保证一号一卡
                </span>
                <span class="inline-flex items-center gap-1.5 rounded-full border border-gray-300/60 dark:border-gray-700 bg-white/80 dark:bg-black/20 px-3.5 py-1 text-xs font-medium text-gray-700 dark:text-gray-200 backdrop-blur-xs">
                  <span class="text-emerald-600">⚡</span> 秒充兑换码快速发货
                </span>
              </div>

            </div>

            <!-- Right Logos: OpenAI & Anthropic (Claude) side-by-side without white background -->
            <div class="shrink-0 flex items-center justify-center gap-6 sm:gap-8 md:gap-10">
              <!-- OpenAI Logo -->
              <div class="flex items-center justify-center group transition-transform hover:scale-105">
                <svg viewBox="0 0 1024 1024" fill="currentColor" class="w-24 h-24 sm:w-28 sm:h-28 md:w-36 md:h-36 text-gray-900 dark:text-white drop-shadow-sm">
                  <path d="M468.0192 30.0544c61.1328 0.8704 118.4256 24.4736 165.4272 71.424 6.912 6.912 12.9536 8.5504 22.2208 7.2192 134.7584-18.9952 253.0304 66.1504 277.8112 200.0384 6.5024 35.2256 4.4032 70.2464-5.5808 104.5504-2.5088 8.3968-0.8704 13.7216 4.5568 20.3264 111.5136 136.8576 44.8 343.7568-125.6448 388.864-9.728 2.5088-13.9776 6.912-17.3568 15.7696-63.3344 164.6592-276.0192 208.9984-400.0256 83.3536-6.1952-6.1952-11.4688-7.5264-19.8656-6.3488-136.704 19.3024-255.1296-65.9968-279.552-201.3696-6.3488-34.7648-3.9936-69.376 6.0416-103.2704 2.5088-8.3968 0.8704-13.7216-4.4032-20.3264-112.0256-136.96-44.1344-345.344 126.8736-389.2224 8.5504-2.2016 12.544-5.888 15.616-13.824C270.9504 90.9312 360.3456 29.6448 468.0192 30.1056z m148.1728 593.1008c-77.312 44.6464-152.1664 87.5008-226.4064 131.072-16.7936 9.8816-31.0784 9.5744-47.872-0.3072-58.1632-34.4576-116.7872-67.8912-175.4112-101.4784-4.2496-2.3552-7.7824-7.2192-14.592-5.1712-7.936 71.5776 14.4384 131.2256 73.216 173.5168 59.1872 42.5472 124.3136 47.872 188.8256 13.7216 65.9968-34.9184 129.7408-74.0864 194.4064-111.5136 3.6864-2.048 7.7824-3.6864 7.7824-9.1136v-90.7264z m-41.216-487.3728c-1.4848-5.1712-5.3248-6.6048-8.2432-8.704-58.9312-39.936-121.6512-46.2336-185.2928-14.8992-64.2048 31.7952-98.4064 86.1696-100.3008 157.9008-1.8944 72.9088-0.3072 145.8176-0.7168 218.7264 0 7.7824 2.6624 11.9296 9.2672 15.4624 21.6576 11.9296 43.008 24.576 64.3584 36.8128 2.6624 1.4848 5.0176 4.7104 9.4208 1.8944 0-86.3232 0.3072-173.056-0.1536-259.84-0.1536-17.8176 6.6048-29.2864 22.1184-37.9904 44.8-25.1904 89.088-51.0976 133.5808-76.9024 18.688-10.752 37.4272-21.8112 55.9616-32.5632z m85.248 341.5552v18.8416c0 81.8688-0.1536 163.9424 0.1536 245.8112 0.1536 16.64-6.0416 28.0064-20.6336 35.9424-20.48 11.2128-40.4992 23.1424-60.672 34.7648-43.4688 25.1904-87.04 50.3808-134.6048 77.9264 4.1472 1.4848 6.1952 1.792 7.68 2.9696 54.6304 39.936 114.2784 49.0496 176.896 24.0128s102.9632-72.6016 110.592-139.6224c9.2672-81.8688 1.8944-164.6592 3.3792-246.9888 0.1536-4.4032-1.8944-6.912-5.4272-8.96-25.0368-14.4384-49.92-28.8768-77.312-44.6464z m-104.1408-111.3088c6.0416 3.6864 10.1376 6.0416 14.1312 8.3968 71.2704 41.216 142.4384 82.6368 214.016 123.4432 15.4624 8.8576 22.1184 20.48 21.9648 38.144-0.4608 68.1984-0.1536 136.3968-0.1536 204.5952 0 5.0176 0.4608 9.8816 0.7168 15.9232 3.6864-1.3312 5.888-2.048 8.0896-2.9696 64.0512-29.7472 102.0928-79.6672 108.1344-149.8112 6.3488-72.6016-22.6816-130.9184-84.992-169.3696-62.72-38.7584-127.5392-74.0864-191.3344-111.2064-5.5808-3.2256-9.728-3.6864-15.616-0.3072-24.1664 14.1312-48.5888 27.8528-74.9568 43.1616z m-338.432-96.8192c-6.6048-0.8704-10.5984 2.2016-14.7456 4.2496-60.9792 30.6176-96.6144 80.128-102.1952 147.8656-5.7344 70.2464 21.504 128 81.152 165.6832 63.9488 40.3456 130.6624 76.4416 196.0448 114.432 3.84 2.2016 7.3728 4.1472 11.776 1.4848 25.344-14.7456 50.688-29.44 77.4656-44.9024-5.5808-3.3792-9.4208-5.888-13.4144-8.2432-71.2704-41.216-142.4384-82.6368-214.016-123.2896-16.0768-9.1136-22.528-21.0432-22.3744-39.3216 0.6144-72.2944 0.3072-144.7936 0.3072-218.0096z m190.1568 130.816c6.1952-3.2256 10.2912-5.3248 14.2848-7.68 70.5536-40.6528 141.2608-80.9984 211.3536-122.4192 16.7936-9.8816 31.0784-10.1376 47.872-0.3072 59.0848 34.9184 118.5792 68.7616 177.92 102.9632 3.5328 2.048 6.5024 5.888 12.544 4.1472 7.7824-70.0928-13.4144-129.3312-70.7072-171.7248-59.4944-44.1856-125.3376-50.2272-191.0272-15.616-65.9968 34.9184-129.7408 74.0864-194.56 111.36-3.6864 2.048-7.68 3.84-7.68 9.2672v90.0096z m208.384 115.9168c0-20.1728-0.4608-36.9664 0.1536-53.6064 0.3072-7.68-2.2016-12.0832-8.96-15.7696-29.44-16.4864-58.7776-33.4336-87.808-50.688-5.5808-3.3792-9.728-3.3792-15.4624-0.1536-29.0304 17.2544-58.3168 34.1504-87.808 50.688-6.7584 3.84-8.8576 8.2432-8.8576 15.9232 0.4608 32.8704 0.6144 65.8432 0 98.6624-0.1536 8.96 3.072 13.7216 10.752 17.664 13.9776 7.2192 27.2384 15.616 40.96 23.3984 17.3568 9.8816 34.7648 27.6992 51.968 27.5456 17.9712-0.1536 35.7888-17.5104 53.4528-27.5456 16.9472-9.728 39.168-16.1792 48.7424-30.7712 9.8816-15.4624 1.792-39.0144 2.816-55.3984z" />
                </svg>
              </div>

              <!-- Anthropic Logo (Gray color, with 即将推出 tag on bottom right) -->
              <div class="relative flex items-center justify-center opacity-85">
                <div class="relative">
                <svg viewBox="0 0 1024 1024" fill="currentColor" class="w-24 h-24 sm:w-28 sm:h-28 md:w-36 md:h-36 text-gray-400 dark:text-gray-500 drop-shadow-xs">
                    <path d="M229.0688 665.4464l183.1424-102.7072 3.1232-8.9088-3.1232-4.9664h-8.9088l-30.6688-1.8432-104.704-2.816-90.7776-3.7888-87.9104-4.7104-22.1696-4.7104-20.7872-27.3408 2.1504-13.6704 18.6368-12.4416 26.624 2.304 58.9824 3.9936 88.3712 6.144 64.1024 3.7888 95.0272 9.8816h15.104l2.1504-6.0928-5.2224-3.7888-3.9936-3.7888-91.4944-61.952-99.0208-65.4848-51.8144-37.7344-28.1088-19.0464-14.1312-17.92-6.144-39.1168 25.4464-28.0064 34.2016 2.304 8.7552 2.3552 34.6624 26.624 74.0352 57.2928L391.2704 380.416l14.1824 11.776 5.632-3.9936 0.7168-2.816-6.3488-10.6496-52.5824-94.9248-56.1152-96.6144-24.9856-40.0384-6.6048-24.0128c-2.5088-9.216-3.8912-18.7392-4.0448-28.2624l29.0304-39.3216 16.0256-5.2224 38.656 5.2224 16.2816 14.1312 24.064 54.8864 38.8608 86.4768L484.352 324.608l17.7152 34.8672 9.4208 32.3072 3.5328 9.8816h6.144v-5.6832l4.9664-66.2016 9.216-81.3056 8.9088-104.5504 3.1232-29.4912 14.592-35.328 28.9792-19.0976 22.6816 10.8544 18.6368 26.5728-2.6112 17.2032-11.1104 71.8336-21.7088 112.64-14.1312 75.3664h8.2432l9.4208-9.3696 38.1952-50.688 64.1024-80.0768 28.3136-31.7952 32.9728-35.072 21.248-16.7424h40.0896l29.4912 43.8272-13.2096 45.2608-41.2672 52.2752-34.2016 44.288-49.0496 65.9456-30.6688 52.7872 2.816 4.2496 7.2704-0.768 110.7968-23.5008 59.8528-10.8544 71.424-12.2368 32.3072 15.0528 3.5328 15.3088-12.7488 31.3344-76.3904 18.8416-89.6 17.92-133.4272 31.5392-1.6384 1.1776 1.8944 2.3552 60.1088 5.6832 25.7024 1.3824h62.9248l117.1968 8.7552 30.6688 20.2752 18.3808 24.7808-3.072 18.8416-47.1552 24.064-63.6416-15.104-148.5824-35.328-50.8928-12.7488h-7.0656v4.2496l42.3936 41.4208 77.824 70.2464 97.3312 90.4192 4.9152 22.4256-12.4928 17.664-13.2096-1.8944-85.5552-64.3072-33.024-28.9792-74.752-62.8736h-4.9664v6.6048l17.2032 25.1904 90.9824 136.6016 4.7104 41.8816-6.6048 13.7216-23.6032 8.2432-25.9072-4.7104-53.2992-74.7008-54.8864-84.0704-44.3392-75.4176-5.4272 3.1232-26.1632 281.4976-12.2368 14.336-28.2624 10.8544-23.552-17.8688-12.4928-28.9792 12.4928-57.2928 15.104-74.6496 12.2368-59.392 11.1104-73.728 6.6048-24.5248-0.4608-1.6384-5.4272 0.7168-55.6544 76.3392-84.5824 114.2784-66.9696 71.5776-16.0768 6.3488-27.8016-14.336 2.6112-25.7024 15.5648-22.8352 92.672-117.8112 55.8592-73.0112 36.096-42.1376-0.256-6.144h-2.1504l-246.1184 159.6928-43.8272 5.6832-18.8928-17.7152 2.3552-28.928 8.96-9.4208 74.0352-50.8928-0.256 0.256z" />
                  </svg>
                  <!-- 即将推出 badge in bottom right -->
                  <span class="absolute -bottom-1 -right-2 sm:-right-4 whitespace-nowrap rounded-md bg-gray-500/15 dark:bg-gray-700/50 border border-gray-400/40 dark:border-gray-600 px-2 py-0.5 text-xs font-medium text-gray-600 dark:text-gray-300 backdrop-blur-xs">
                    即将推出
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- ==================== FEATURED SERVICES ==================== -->
    <section id="featured-services" class="relative z-10 py-8 md:py-12">
      <div class="max-w-6xl mx-auto px-4">
        <div class="mb-6 flex items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <h2 class="shrink-0 text-xl sm:text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
              精选服务
            </h2>
            <a
              v-if="telegramContactUrl"
              :href="telegramContactUrl"
              target="_blank"
              rel="noopener noreferrer"
              aria-label="通过 Telegram 联系我们"
              title="通过 Telegram 联系我们"
              class="inline-flex shrink-0 items-center gap-2 rounded-full border border-sky-200 bg-sky-50/80 px-3 py-1.5 text-xs font-semibold text-sky-700 transition hover:border-sky-300 hover:bg-sky-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sky-500 focus-visible:ring-offset-2 dark:border-sky-900 dark:bg-sky-950/40 dark:text-sky-300 dark:hover:bg-sky-950"
            >
              <svg class="h-4 w-4" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221l-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.446 1.394c-.14.18-.357.295-.6.295-.002 0-.003 0-.005 0l.213-3.054 5.56-5.022c.24-.213-.054-.334-.373-.121l-6.869 4.326-2.96-.924c-.64-.203-.658-.64.135-.954l11.566-4.458c.538-.196 1.006.128.832.941z" />
              </svg>
              <span>联系我们</span>
            </a>
          </div>
          <span class="text-xs text-muted-foreground">官方独立直充通道</span>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-5">
          <div
            v-for="item in visibleFeaturedList"
            :key="item.slug"
            class="flex flex-col justify-between rounded-2xl border bg-card p-4 md:p-5 shadow-xs transition hover:shadow-md hover:border-emerald-500/40"
          >
            <!-- Card Header -->
            <div class="space-y-3">
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
                <span class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300 text-[10px] font-bold">AI</span>
                <span>{{ item.slug === 'openai-sms-activate' ? 'OPENAI · 短信接码' : 'CHATGPT · 订阅服务' }}</span>
              </div>

              <div>
                <div class="flex items-start justify-between gap-2">
                <h3 class="min-w-0 text-base font-bold text-gray-900 dark:text-white">
                  {{ item.name }}
                </h3>
                <div class="shrink-0 text-right font-bold text-gray-900 dark:text-white md:hidden">
                  <template v-if="item.price">¥{{ item.price }}</template>
                  <span v-else class="text-[11px] font-medium text-muted-foreground">{{ productsLoadError || !productsLoading ? '价格暂不可用' : '正在加载价格…' }}</span>
                </div>
                </div>
                <p class="mt-1 hidden text-xs leading-relaxed text-muted-foreground md:block">
                  {{ t(item.descriptionKey) }}
                </p>
              </div>

              <details class="group rounded-xl border border-border/80 bg-muted/30 md:hidden">
                <summary class="flex min-h-9 cursor-pointer list-none items-center justify-between gap-2 px-3 text-xs font-semibold text-foreground [&::-webkit-details-marker]:hidden">
                  <span>{{ item.slug === 'openai-sms-activate' ? t('homeMobile.smsCardDetails') : t('homeMobile.cardDetails') }}</span>
                  <ChevronDown class="h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform group-open:rotate-180" />
                </summary>
                <div class="space-y-2 border-t px-3 py-2.5">
                  <p class="text-xs leading-relaxed text-muted-foreground">{{ t(item.descriptionKey) }}</p>
                  <div class="rounded-lg border border-amber-200/80 bg-amber-50/70 p-2.5 text-[11px] leading-snug text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200">
                    <div class="flex items-center gap-1 font-semibold">
                      <AlertCircle class="h-3 w-3 shrink-0 text-amber-600" />
                      <span>{{ t(item.requirementTitleKey) }}</span>
                    </div>
                    <div class="mt-0.5 text-amber-800/90 dark:text-amber-300/90">
                      {{ t(item.requirementKey) }}
                    </div>
                  </div>
                </div>
              </details>

              <!-- Requirement Box -->
              <div class="hidden rounded-xl border border-amber-200/80 bg-amber-50/70 p-2.5 text-[11px] leading-snug text-amber-900 dark:border-amber-900/50 dark:bg-amber-950/30 dark:text-amber-200 md:block">
                <div class="font-semibold flex items-center gap-1">
                  <AlertCircle class="h-3 w-3 shrink-0 text-amber-600" />
                  <span>{{ t(item.requirementTitleKey) }}</span>
                </div>
                <div class="mt-0.5 text-amber-800/90 dark:text-amber-300/90">
                  {{ t(item.requirementKey) }}
                </div>
              </div>

              <!-- Flow Link -->
              <div class="flex items-center justify-between gap-2 md:block">
                <button
                  type="button"
                  @click="openDesktopDetails(item)"
                  class="hidden items-center gap-1 text-xs font-medium text-emerald-600 hover:text-emerald-700 dark:text-emerald-400 md:inline-flex"
                >
                  <span>查看详细运行流程</span>
                  <ArrowUpRight class="h-3 w-3" />
                </button>
                <button type="button" class="group inline-flex min-h-9 min-w-0 items-center gap-2 rounded-xl border border-emerald-200/80 bg-emerald-50/70 px-3 text-left text-[11px] font-semibold text-emerald-800 shadow-xs transition hover:border-emerald-300 hover:bg-emerald-50 active:scale-[0.98] dark:border-emerald-900 dark:bg-emerald-950/40 dark:text-emerald-300 md:hidden" @click="openMobileTerms(item)"><span class="truncate">{{ t('homeMobile.terms') }}</span><ChevronDown class="h-3.5 w-3.5 shrink-0 -rotate-90 transition-transform group-hover:translate-x-0.5" /></button>
                <button type="button" :disabled="!productsLoaded || !products.some((product) => product.slug === item.slug)" class="shrink-0 rounded-lg bg-gray-900 px-3 py-2 text-[11px] font-semibold text-white disabled:opacity-50 dark:bg-white dark:text-gray-900 md:hidden" @click="handleQuickBuy(item.slug)">{{ t('homeMobile.buy') }}</button>
              </div>
            </div>

            <!-- Card Bottom -->
            <div class="mt-5 hidden items-center justify-between gap-2 border-t pt-4 md:flex">
              <div>
                <div class="text-lg font-bold text-gray-900 dark:text-white">
                  <template v-if="item.price">¥{{ item.price }}<span class="text-xs font-normal text-muted-foreground"> / 份</span></template>
                  <span v-else class="text-sm font-medium text-muted-foreground">{{ productsLoadError || !productsLoading ? '价格暂不可用' : '正在加载价格…' }}</span>
                </div>
              </div>

              <button
                type="button"
                :disabled="!productsLoaded || !products.some((product) => product.slug === item.slug)"
                @click="handleQuickBuy(item.slug)"
                class="rounded-xl bg-gray-900 px-3.5 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-gray-800 active:scale-95 dark:bg-white dark:text-gray-900 dark:hover:bg-gray-100 shrink-0 disabled:cursor-not-allowed disabled:opacity-50"
              >
                立即购买
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <section v-if="hasAdditionalProducts" class="relative z-10 py-8 md:py-10 border-t">
      <div class="max-w-6xl mx-auto px-4">
        <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-xl font-bold text-foreground">更多商品</h2>
            <p class="mt-1 text-xs text-muted-foreground">商品价格以当前页面显示和订单结算为准。</p>
          </div>
          <label class="relative block w-full sm:max-w-xs">
            <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input v-model="additionalProductQuery" type="search" class="h-10 w-full rounded-xl border bg-card pl-9 pr-3 text-sm" placeholder="搜索商品" />
          </label>
        </div>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <article v-for="product in filteredAdditionalProducts" :key="product.id" class="flex items-center justify-between gap-4 rounded-xl border bg-card p-4">
            <div class="min-w-0">
              <h3 class="truncate text-sm font-semibold">{{ getLocalizedText(product.title) }}</h3>
              <p class="mt-1 line-clamp-2 text-xs text-muted-foreground">{{ getLocalizedText(product.description) }}</p>
              <div class="mt-2 font-bold text-primary">{{ formatProductPrice(product.price_amount) }}</div>
            </div>
            <button type="button" :disabled="!isPriceAvailable(product.price_amount)" class="shrink-0 rounded-lg bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground disabled:cursor-not-allowed disabled:opacity-50" @click="handleQuickBuy(product.slug)">立即购买</button>
          </article>
        </div>
        <p v-if="additionalProductQuery && !filteredAdditionalProducts.length" class="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">没有找到匹配的商品。</p>
      </div>
    </section>

    <!-- Desktop product details dialog -->
    <Teleport to="body">
      <div v-if="desktopDetailsItem" class="fixed inset-0 z-[80] hidden items-center justify-center bg-slate-950/50 p-4 backdrop-blur-[3px] md:flex" @click.self="desktopDetailsItem = null">
        <section role="dialog" aria-modal="true" :aria-label="`${desktopDetailsItem.name} 详细运行流程`" class="max-h-[88vh] w-full max-w-5xl overflow-y-auto rounded-3xl border bg-background p-5 text-foreground shadow-2xl md:p-8">
          <div class="sticky top-0 z-10 -mx-5 -mt-5 mb-5 flex items-center justify-between gap-4 border-b bg-background/95 px-5 py-4 backdrop-blur md:-mx-8 md:-mt-8 md:px-8">
            <div class="min-w-0">
              <p class="truncate text-xs text-muted-foreground">{{ desktopDetailsItem.name }}</p>
              <h2 class="mt-1 text-lg font-bold">详细运行流程</h2>
            </div>
            <button type="button" class="shrink-0 rounded-lg border px-3 py-2 text-xs font-semibold text-muted-foreground transition hover:bg-muted" aria-label="关闭详细说明" @click="desktopDetailsItem = null">关闭</button>
          </div>
          <div class="space-y-6">

        <!-- 1. ChatGPT Plus 详细说明 -->
        <div v-if="desktopDetailsItem?.slug === 'chatgpt-plus-1m' && productsLoaded && products.some((product) => product.slug === 'chatgpt-plus-1m')" id="pkg-chatgpt-plus-1m" class="rounded-3xl border bg-card p-6 md:p-8 space-y-6 shadow-xs">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-md text-[11px] font-semibold bg-emerald-100 text-emerald-800 dark:bg-emerald-950 dark:text-emerald-300">
                PLUS · 1个月
              </span>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white mt-1.5">
                ChatGPT Plus · 1个月
              </h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                自有个人账户开通 · 30天订阅质保
              </p>
            </div>
            <button
              type="button"
              @click="handleQuickBuy('chatgpt-plus-1m')"
              class="self-start sm:self-auto rounded-xl bg-gray-900 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-gray-800 dark:bg-white dark:text-gray-900"
            >
              立即购买
            </button>
          </div>

          <!-- 3 Steps -->
          <div>
            <div class="text-xs font-semibold text-gray-500 mb-3 uppercase tracking-wider">开通流程</div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-600 text-[10px] text-white">1</span>
                  <span>选择套餐并付款</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">选择套餐后，请在结算页选择已开放的支付方式；支付宝通道审批通过并开放后，也可选择支付宝。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-600 text-[10px] text-white">2</span>
                  <span>按订单提示提交 Session</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">支付确认后，请按订单详情提示提交 ChatGPT Session。请勿向任何人提供账户密码。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-emerald-600 text-[10px] text-white">3</span>
                  <span>客服核验并开通</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">客服核验订单信息后处理开通；完成后订单状态会更新。</p>
              </div>
            </div>
          </div>

          <!-- Conditions & Warranty -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
            <div class="space-y-3">
              <div class="text-xs font-bold text-foreground">📝 使用条件</div>
              <ol class="space-y-2 text-xs text-muted-foreground">
                <li v-for="(line, index) in purchaseConditions.plus" :key="line" class="flex gap-2"><span class="font-mono text-gray-400">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
              </ol>
            </div>
            <div class="space-y-3 border-t pt-4 md:border-l md:border-t-0 md:pl-6 md:pt-0">
              <div class="text-xs font-bold text-foreground">🛡️ 质保订阅30天</div>
              <p v-for="line in purchaseWarranty" :key="line" class="text-xs leading-relaxed text-muted-foreground">{{ line }}</p>
            </div>
          </div>
        </div>

        <!-- 2. ChatGPT Pro 5x 详细说明 -->
        <div v-if="desktopDetailsItem?.slug === 'chatgpt-pro-5x-1m' && productsLoaded && products.some((product) => product.slug === 'chatgpt-pro-5x-1m')" id="pkg-chatgpt-pro-5x-1m" class="rounded-3xl border bg-card p-6 md:p-8 space-y-6 shadow-xs">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-md text-[11px] font-semibold bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-300">
                PRO 5X · 1个月
              </span>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white mt-1.5">
                ChatGPT Pro 5x · 1个月
              </h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                自有个人账户开通 · 30天订阅质保
              </p>
            </div>
            <button
              type="button"
              @click="handleQuickBuy('chatgpt-pro-5x-1m')"
              class="self-start sm:self-auto rounded-xl bg-gray-900 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-gray-800 dark:bg-white dark:text-gray-900"
            >
              立即购买
            </button>
          </div>

          <!-- 3 Steps -->
          <div>
            <div class="text-xs font-semibold text-gray-500 mb-3 uppercase tracking-wider">开通流程</div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-blue-600 text-[10px] text-white">1</span>
                  <span>选择套餐并付款</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">选择套餐后，请在结算页选择已开放的支付方式；支付宝通道审批通过并开放后，也可选择支付宝。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-blue-600 text-[10px] text-white">2</span>
                  <span>按订单提示提交 Session</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">支付确认后，请按订单详情提示提交 ChatGPT Session。请勿向任何人提供账户密码。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-blue-600 text-[10px] text-white">3</span>
                  <span>客服核验并开通</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">客服核验订单信息后处理开通；完成后订单状态会更新。</p>
              </div>
            </div>
          </div>

          <!-- Conditions & Warranty -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
            <div class="space-y-3">
              <div class="text-xs font-bold text-foreground">📝 使用条件</div>
              <ol class="space-y-2 text-xs text-muted-foreground">
                <li v-for="(line, index) in purchaseConditions.pro" :key="line" class="flex gap-2"><span class="font-mono text-gray-400">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
              </ol>
            </div>
            <div class="space-y-3 border-t pt-4 md:border-l md:border-t-0 md:pl-6 md:pt-0">
              <div class="text-xs font-bold text-foreground">🛡️ 质保订阅30天</div>
              <p v-for="line in purchaseWarranty" :key="line" class="text-xs leading-relaxed text-muted-foreground">{{ line }}</p>
            </div>
          </div>
        </div>

        <!-- 3. ChatGPT Pro 20x 详细说明 -->
        <div v-if="desktopDetailsItem?.slug === 'chatgpt-pro-20x-1m' && productsLoaded && products.some((product) => product.slug === 'chatgpt-pro-20x-1m')" id="pkg-chatgpt-pro-20x-1m" class="rounded-3xl border bg-card p-6 md:p-8 space-y-6 shadow-xs">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-md text-[11px] font-semibold bg-purple-100 text-purple-800 dark:bg-purple-950 dark:text-purple-300">
                PRO 20X · 1个月
              </span>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white mt-1.5">
                ChatGPT Pro 20x · 1个月
              </h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                正规充值服务 · 付款后提交 Session
              </p>
            </div>
            <button
              type="button"
              @click="handleQuickBuy('chatgpt-pro-20x-1m')"
              class="self-start sm:self-auto rounded-xl bg-gray-900 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-gray-800 dark:bg-white dark:text-gray-900"
            >
              立即购买
            </button>
          </div>

          <!-- Notice Box -->
          <div class="rounded-2xl border border-purple-200/80 bg-purple-50/60 dark:border-purple-900/40 dark:bg-purple-950/20 p-4 space-y-1">
            <div class="text-xs font-bold text-purple-900 dark:text-purple-300">
              ChatGPT Pro 20x 充值服务须知
            </div>
            <p class="text-xs text-purple-800/80 dark:text-purple-300/80 leading-relaxed">
              官方正规代充 Pro 20x 充值服务，如当前账号已有 Plus, Pro 5x，支持升级差额充值。
            </p>
          </div>

          <!-- 3 Steps -->
          <div>
            <div class="text-xs font-semibold text-gray-500 mb-3 uppercase tracking-wider">开通流程</div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-purple-600 text-[10px] text-white">1</span>
                  <span>选择套餐并付款</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">选择套餐后，请在结算页选择已开放的支付方式；支付宝通道审批通过并开放后，也可选择支付宝。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-purple-600 text-[10px] text-white">2</span>
                  <span>按订单提示提交 Session</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">支付确认后，请按订单详情提示提交 ChatGPT Session。请勿向任何人提供账户密码。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-purple-600 text-[10px] text-white">3</span>
                  <span>客服核验并开通</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">客服核验订单信息后处理开通；完成后订单状态会更新。</p>
              </div>
            </div>
          </div>

          <!-- Restrictions & Refund -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
            <div class="space-y-3">
              <div class="flex items-center gap-1.5 text-xs font-bold text-gray-900 dark:text-white">
                <span class="text-base">⚠️</span>
                <span>账号下单限制</span>
              </div>
              <ul class="space-y-2 text-xs text-muted-foreground">
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">01</span>
                  <span>当前账号没有会员，或当前为菲区 Plus、Pro 5x，可升级。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">02</span>
                  <span>曾购买非正常 Pro（0 元购 Pro、低价黑冲账号）请勿下单，否则可能订阅掉线、触发邮件欺诈，建议更换账号再提交。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">03</span>
                  <span>本服务仅限 Codex Pro 订阅正价消费。</span>
                </li>
              </ul>
            </div>

            <div class="space-y-3 border-t md:border-t-0 md:border-l pt-4 md:pt-0 md:pl-6">
              <div class="flex items-center gap-1.5 text-xs font-bold text-gray-900 dark:text-white">
                <span class="text-base">📜</span>
                <span>退款相关条款</span>
              </div>
              <ul class="space-y-2 text-xs text-muted-foreground">
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">01</span>
                  <span>仅因平台 / 服务自身故障问题可申请退款，其余情况一律不予退款。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">02</span>
                  <span>因用户违规（邮件欺诈、网络滥用、违反平台条款等）导致问题，无退款资格。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">03</span>
                  <span>账号封禁、KYC 未通过、未绑定手机号、自行人工取消订阅等个人原因，不予退款。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">04</span>
                  <span>若出现账号封禁，需自行联系 OpenAI 官方客服处理退款，本平台不代办、不担保结果。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">05</span>
                  <span>若符合退款条件，退款将扣除 20% 手续费，剩余金额按实际情况返还。</span>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <!-- 4. ChatGPT Pro 50x 详细说明 -->
        <div v-if="desktopDetailsItem?.slug === 'chatgpt-pro-50x-1m' && productsLoaded && products.some((product) => product.slug === 'chatgpt-pro-50x-1m')" id="pkg-chatgpt-pro-50x-1m" class="rounded-3xl border bg-card p-6 md:p-8 space-y-6 shadow-xs">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b">
            <div>
              <span class="inline-block px-2.5 py-0.5 rounded-md text-[11px] font-semibold bg-rose-100 text-rose-800 dark:bg-rose-950 dark:text-rose-300">
                PRO 50X · 1个月
              </span>
              <h3 class="text-xl font-bold text-gray-900 dark:text-white mt-1.5">
                ChatGPT Pro 50x · 1个月
              </h3>
              <p class="text-xs text-muted-foreground mt-0.5">
                顶级算力充值 · 付款后提交 Session
              </p>
            </div>
            <button
              type="button"
              @click="handleQuickBuy('chatgpt-pro-50x-1m')"
              class="self-start sm:self-auto rounded-xl bg-gray-900 px-4 py-2 text-xs font-semibold text-white shadow-xs transition hover:bg-gray-800 dark:bg-white dark:text-gray-900"
            >
              立即购买
            </button>
          </div>

          <!-- Notice Box -->
          <div class="rounded-2xl border border-rose-200/80 bg-rose-50/60 dark:border-rose-900/40 dark:bg-rose-950/20 p-4 space-y-1">
            <div class="text-xs font-bold text-rose-900 dark:text-rose-300">
              ChatGPT Pro 50x 充值服务须知
            </div>
            <p class="text-xs text-rose-800/80 dark:text-rose-300/80 leading-relaxed">
              官方正规代充 Pro 50x 顶级算力充值服务，如当前账号已有 Plus, Pro 5x, Pro 20x，支持升级差额充值。
            </p>
          </div>

          <!-- 3 Steps -->
          <div>
            <div class="text-xs font-semibold text-gray-500 mb-3 uppercase tracking-wider">开通流程</div>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-rose-600 text-[10px] text-white">1</span>
                  <span>选择套餐并付款</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">选择套餐后，请在结算页选择已开放的支付方式；支付宝通道审批通过并开放后，也可选择支付宝。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-rose-600 text-[10px] text-white">2</span>
                  <span>按订单提示提交 Session</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">支付确认后，请按订单详情提示提交 ChatGPT Session。请勿向任何人提供账户密码。</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-1">
                <div class="flex items-center gap-2 text-xs font-bold text-gray-900 dark:text-white">
                  <span class="flex h-5 w-5 items-center justify-center rounded-full bg-rose-600 text-[10px] text-white">3</span>
                  <span>客服核验并开通</span>
                </div>
                <p class="text-xs text-muted-foreground pl-7">客服核验订单信息后处理开通；完成后订单状态会更新。</p>
              </div>
            </div>
          </div>

          <!-- Restrictions & Refund -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
            <div class="space-y-3">
              <div class="flex items-center gap-1.5 text-xs font-bold text-gray-900 dark:text-white">
                <span class="text-base">⚠️</span>
                <span>账号下单限制</span>
              </div>
              <ul class="space-y-2 text-xs text-muted-foreground">
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">01</span>
                  <span>当前账号没有会员，或当前为菲区 Plus、Pro 5x、Pro 20x，可升级。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">02</span>
                  <span>曾购买非正常 Pro（0 元购 Pro、低价黑冲账号）请勿下单，否则可能订阅掉线、触发邮件欺诈，建议更换账号再提交。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">03</span>
                  <span>本服务仅限 Codex Pro 订阅正价消费。</span>
                </li>
              </ul>
            </div>

            <div class="space-y-3 border-t md:border-t-0 md:border-l pt-4 md:pt-0 md:pl-6">
              <div class="flex items-center gap-1.5 text-xs font-bold text-gray-900 dark:text-white">
                <span class="text-base">📜</span>
                <span>退款相关条款</span>
              </div>
              <ul class="space-y-2 text-xs text-muted-foreground">
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">01</span>
                  <span>仅因平台 / 服务自身故障问题可申请退款，其余情况一律不予退款。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">02</span>
                  <span>因用户违规（邮件欺诈、网络滥用、违反平台条款等）导致问题，无退款资格。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">03</span>
                  <span>账号封禁、KYC 未通过、未绑定手机号、自行人工取消订阅等个人原因，不予退款。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">04</span>
                  <span>若出现账号封禁，需自行联系 OpenAI 官方客服处理退款，本平台不代办、不担保结果。</span>
                </li>
                <li class="flex items-start gap-2">
                  <span class="font-mono text-gray-400">05</span>
                  <span>若符合退款条件，退款将扣除 20% 手续费，剩余金额按实际情况返还。</span>
                </li>
              </ul>
            </div>
          </div>
        </div>

          <div v-if="desktopDetailsItem?.slug === goProductSlug" class="space-y-5 rounded-3xl border bg-card p-6 shadow-xs md:p-8">
            <div class="flex items-center justify-between gap-3 border-b pb-4">
              <div>
                <h3 class="text-lg font-bold">{{ t('homeMobile.goTitle') }}</h3>
                <p class="mt-1 text-sm text-muted-foreground">{{ t('homeMobile.goDescription') }}</p>
              </div>
              <button type="button" class="shrink-0 rounded-xl bg-gray-900 px-4 py-2 text-xs font-semibold text-white dark:bg-white dark:text-gray-900" @click="handleQuickBuy(goProductSlug)">{{ t('homeMobile.buy') }}</button>
            </div>
            <ol class="grid gap-3 md:grid-cols-3">
              <li v-for="(line, index) in goFlowSteps" :key="index" class="rounded-2xl border bg-muted/30 p-4 text-sm leading-relaxed text-muted-foreground"><span class="mb-2 block font-semibold text-emerald-700 dark:text-emerald-300">{{ String(index + 1).padStart(2, '0') }}</span>{{ line }}</li>
            </ol>
            <div class="grid gap-4 md:grid-cols-2">
              <div class="rounded-2xl border bg-muted/30 p-4">
                <h4 class="mb-3 text-sm font-bold">{{ t('homeMobile.conditions') }}</h4>
                <p v-for="line in goConditions" :key="line" class="mt-2 text-xs leading-relaxed text-muted-foreground">{{ line }}</p>
              </div>
              <div class="rounded-2xl border bg-muted/30 p-4">
                <h4 class="mb-3 text-sm font-bold">{{ t('homeMobile.warranty') }}</h4>
                <p v-for="line in purchaseWarranty" :key="line" class="mt-2 text-xs leading-relaxed text-muted-foreground">{{ line }}</p>
              </div>
            </div>
          </div>

          <div v-if="desktopDetailsItem?.slug === 'openai-sms-activate'" class="rounded-3xl border bg-card p-6 shadow-xs md:p-8">
            <div class="flex items-start gap-3 border-b pb-5">
              <span class="inline-flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-xs font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">SMS</span>
              <div>
                <h3 class="text-lg font-bold">{{ t('homeMobile.smsDetailsTitle') }}</h3>
                <p class="mt-1 text-sm text-muted-foreground">{{ t('homeMobile.smsDetailsDescription') }}</p>
              </div>
            </div>
            <ol class="mt-5 grid gap-3 md:grid-cols-3">
              <li class="rounded-2xl border bg-muted/30 p-4">
                <span class="text-xs font-semibold text-emerald-700 dark:text-emerald-300">01 · {{ t('homeMobile.smsStep1Title') }}</span>
                <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ t('homeMobile.smsStep1Description') }}</p>
              </li>
              <li class="rounded-2xl border bg-muted/30 p-4">
                <span class="text-xs font-semibold text-emerald-700 dark:text-emerald-300">02 · {{ t('homeMobile.smsStep2Title') }}</span>
                <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ t('homeMobile.smsStep2Description') }}</p>
              </li>
              <li class="rounded-2xl border bg-muted/30 p-4">
                <span class="text-xs font-semibold text-emerald-700 dark:text-emerald-300">03 · {{ t('homeMobile.smsStep3Title') }}</span>
                <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ t('homeMobile.smsStep3Description') }}</p>
              </li>
            </ol>
            <div class="mt-5 rounded-2xl border border-sky-100 bg-sky-50/55 p-4 dark:border-sky-900/50 dark:bg-sky-950/20">
              <h4 class="flex items-center gap-2 text-sm font-bold text-sky-900 dark:text-sky-200"><span class="h-1.5 w-1.5 rounded-full bg-sky-500"></span>{{ t('homeMobile.smsWarranty') }}</h4>
              <ol class="mt-3 space-y-2 text-sm leading-relaxed text-foreground/75">
                <li v-for="(line, index) in smsFulfillmentGuarantee" :key="index" class="flex items-start gap-2.5"><span class="mt-px inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-white text-[10px] font-bold text-sky-700 shadow-xs dark:bg-sky-900/70 dark:text-sky-200">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
              </ol>
            </div>
          </div>
          </div>
        </section>
      </div>
    </Teleport>

    <!-- ==================== LATEST POSTS (IF ENABLED) ==================== -->
    <section v-if="latestSectionVisible && posts.length > 0" class="relative z-10 py-10 border-t">
      <div class="max-w-6xl mx-auto px-4">
        <div class="mb-6 flex items-end justify-between gap-4">
          <div>
            <h2 class="text-xl font-bold tracking-tight text-gray-900 dark:text-white">{{ t('home.latest.title') }}</h2>
            <p class="mt-1 text-xs text-muted-foreground">{{ t('home.latest.description') }}</p>
          </div>
          <div class="flex items-center gap-3 text-xs">
            <router-link v-if="blogEnabled" to="/blog" class="text-muted-foreground transition-colors hover:text-foreground">{{ t('nav.blog') }}</router-link>
            <router-link v-if="noticeEnabled" to="/notice" class="text-muted-foreground transition-colors hover:text-foreground">{{ t('nav.notice') }}</router-link>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <article
            v-for="post in posts"
            :key="post.id"
            class="cursor-pointer rounded-2xl border bg-card p-5 shadow-xs transition hover:shadow-md"
            @click="goToPost(post.slug)"
          >
            <div class="mb-2 text-xs text-muted-foreground">{{ formatDate(post.published_at) }}</div>
            <h3 class="line-clamp-2 text-sm font-semibold">{{ getLocalizedText(post.title) }}</h3>
            <p class="mt-2 line-clamp-2 text-xs text-muted-foreground">{{ getLocalizedText(post.summary) }}</p>
            <div class="mt-3 text-xs font-medium text-emerald-600">{{ t('blog.readMore') }}</div>
          </article>
        </div>
      </div>
    </section>

    <!-- Quick Buy Dialog -->
    <Teleport to="body">
      <div v-if="mobileTermsItem" class="fixed inset-0 z-[80] flex items-end justify-center overscroll-contain bg-slate-950/45 backdrop-blur-[3px] md:hidden" @click.self="mobileTermsItem = null">
        <section role="dialog" aria-modal="true" :aria-label="t('homeMobile.terms')" data-scroll-lock-scrollable class="max-h-[86dvh] w-full overflow-y-auto overscroll-contain rounded-t-[1.6rem] border border-white/70 bg-background px-5 pb-[calc(1.25rem+env(safe-area-inset-bottom))] pt-3 text-foreground shadow-[0_-16px_50px_-24px_rgba(15,23,42,0.35)]">
          <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-muted-foreground/25"></div>
          <div class="sticky top-0 z-10 -mx-5 flex items-center justify-between gap-3 border-b border-border/70 bg-background/95 px-5 pb-4 pt-1 backdrop-blur">
            <div class="min-w-0">
              <p class="truncate text-xs font-medium text-muted-foreground">{{ mobileTermsItem.name }}</p>
              <h2 class="mt-1 text-lg font-bold tracking-tight">{{ t('homeMobile.terms') }}</h2>
            </div>
            <button type="button" class="rounded-full border bg-muted/60 px-3 py-1.5 text-xs font-semibold text-muted-foreground transition hover:bg-muted" :aria-label="t('homeMobile.close')" @click="mobileTermsItem = null">{{ t('homeMobile.close') }}</button>
          </div>
          <div class="mt-4 rounded-2xl border border-emerald-100 bg-emerald-50/55 p-4 dark:border-emerald-900/50 dark:bg-emerald-950/20">
            <h3 class="flex items-center gap-2 text-sm font-bold text-emerald-900 dark:text-emerald-200"><span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>{{ t('homeMobile.conditions') }}</h3>
            <ol class="mt-3 space-y-3 text-[13px] leading-relaxed text-foreground/75">
              <li v-for="(line, index) in mobileTermsConditions" :key="index" class="flex items-start gap-2.5"><span class="mt-px inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-white text-[10px] font-bold text-emerald-700 shadow-xs dark:bg-emerald-900/70 dark:text-emerald-200">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
            </ol>
          </div>
          <div class="mt-3 rounded-2xl border border-sky-100 bg-sky-50/55 p-4 dark:border-sky-900/50 dark:bg-sky-950/20">
            <h3 class="flex items-center gap-2 text-sm font-bold text-sky-900 dark:text-sky-200"><span class="h-1.5 w-1.5 rounded-full bg-sky-500"></span>{{ mobileTermsWarrantyTitle }}</h3>
            <ol class="mt-3 space-y-3 text-[13px] leading-relaxed text-foreground/75">
              <li v-for="(line, index) in mobileTermsWarranty" :key="index" class="flex items-start gap-2.5"><span class="mt-px inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-white text-[10px] font-bold text-sky-700 shadow-xs dark:bg-sky-900/70 dark:text-sky-200">{{ String(index + 1).padStart(2, '0') }}</span><span>{{ line }}</span></li>
            </ol>
          </div>
        </section>
      </div>
    </Teleport>

    <ProductQuickBuy
      v-if="quickBuyProduct"
      :product="quickBuyProduct"
      :visible="quickBuyVisible"
      @update:visible="quickBuyVisible = $event"
    />

    <!-- Announcement Dialog -->
    <AnnouncementModal
      v-if="activeAnnouncement"
      :announcement="activeAnnouncement"
      :visible="announcementVisible"
      @update:visible="announcementVisible = $event"
    />
  </div>
</template>

<script setup lang="ts">
import PlusHalfCheck from '../components/PlusHalfCheck.vue'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  AlertCircle,
  ArrowUpRight,
  ChevronDown,
  Search,
} from 'lucide-vue-next'
import { postAPI, productAPI } from '../api'
import { useLocalized } from '../composables/useProduct'
import { useAppStore } from '../stores/app'
import ProductQuickBuy from '../components/ProductQuickBuy.vue'
import { purchaseConditions, purchaseWarranty } from '../utils/purchaseTerms'
import AnnouncementModal from '../components/AnnouncementModal.vue'
import { useAnnouncement, type HomeAnnouncement } from '../composables/useAnnouncement'

const router = useRouter()
const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const appStore = useAppStore()
const telegramContactUrl = computed(() => String(appStore.config?.contact?.telegram || '').trim())

const goProductSlug = 'upstream-1-1109423947-1791826450925'

// Featured Products configuration
const featuredList = ref([
  {
    slug: 'chatgpt-plus-1m',
    targetId: 'pkg-chatgpt-plus-1m',
    name: 'ChatGPT Plus · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.plusRequirement',
    price: '168.00',
  },
  {
    slug: goProductSlug,
    targetId: 'pkg-chatgpt-go',
    name: 'ChatGPT Go',
    descriptionKey: 'productCardCopy.goDescription',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.goRequirement',
    price: '56.00',
  },
  {
    slug: 'chatgpt-pro-5x-1m',
    targetId: 'pkg-chatgpt-pro-5x-1m',
    name: 'ChatGPT Pro 5x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro5Requirement',
    price: '899.00',
  },
  {
    slug: 'chatgpt-pro-20x-1m',
    targetId: 'pkg-chatgpt-pro-20x-1m',
    name: 'ChatGPT Pro 20x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro20Requirement',
    price: '1688.00',
  },
  {
    slug: 'chatgpt-pro-50x-1m',
    targetId: 'pkg-chatgpt-pro-50x-1m',
    name: 'ChatGPT Pro 50x · 1个月',
    descriptionKey: 'productCardCopy.subscriptionWarranty',
    requirementTitleKey: 'productCardCopy.accountRequirement',
    requirementKey: 'productCardCopy.pro50Requirement',
    price: '3899.00',
  },
  {
    slug: 'openai-sms-activate',
    targetId: 'pkg-openai-sms-activate',
    name: 'OpenAI 接码台',
    descriptionKey: 'productCardCopy.smsDescription',
    requirementTitleKey: 'productCardCopy.deliveryRule',
    requirementKey: 'productCardCopy.smsRequirement',
    price: '19.00',
  },
])
type FeaturedItem = (typeof featuredList.value)[number]
const mobileTermsItem = ref<FeaturedItem | null>(null)
const desktopDetailsItem = ref<FeaturedItem | null>(null)
const smsFulfillmentGuarantee = computed(() => [
  t('homeMobile.smsGuarantee1'),
  t('homeMobile.smsGuarantee2'),
  t('homeMobile.smsGuarantee3'),
])
const goConditions = computed(() => [t('homeMobile.goCondition1'), t('homeMobile.goCondition2'), t('homeMobile.goCondition3')])
const goFlowSteps = computed(() => [t('homeMobile.goStep1'), t('homeMobile.goStep2'), t('homeMobile.goStep3')])
const mobileTermsConditions = computed(() => {
  const item = mobileTermsItem.value
  if (!item) return []
  if (item.slug === 'openai-sms-activate') {
    return [t('homeMobile.smsCondition1'), t('homeMobile.smsCondition2')]
  }
  if (item.slug === goProductSlug) return goConditions.value
  if (item.slug === 'chatgpt-plus-1m') return purchaseConditions.plus
  if (item.slug === 'chatgpt-pro-5x-1m') return purchaseConditions.pro
  return [t(item.requirementKey) + '。', ...purchaseConditions.pro.slice(1)]
})
const mobileTermsWarranty = computed(() => mobileTermsItem.value?.slug === 'openai-sms-activate' ? smsFulfillmentGuarantee.value : purchaseWarranty)
const mobileTermsWarrantyTitle = computed(() => mobileTermsItem.value?.slug === 'openai-sms-activate' ? t('homeMobile.smsWarranty') : t('homeMobile.warranty'))
const openMobileTerms = (item: FeaturedItem) => {
  mobileTermsItem.value = item
}
const openDesktopDetails = (item: FeaturedItem) => {
  desktopDetailsItem.value = item
}

const products = ref<any[]>([])
const productsLoading = ref(true)
const productsLoaded = ref(false)
const productsLoadError = ref('')
const additionalProductQuery = ref('')
const featuredSlugs = computed(() => new Set(featuredList.value.map((item) => item.slug)))
const visibleFeaturedList = computed(() => {
  if (!productsLoaded.value) return []
  const featuredBySlug = new Map(featuredList.value.map((item) => [item.slug, item]))
  return products.value
    .map((product) => featuredBySlug.get(String(product.slug || '')))
    .filter((item): item is FeaturedItem => Boolean(item))
})
const hasAdditionalProducts = computed(() => products.value.some((product) => !featuredSlugs.value.has(product.slug)))
const filteredAdditionalProducts = computed(() => {
  const query = additionalProductQuery.value.trim().toLocaleLowerCase()
  return products.value.filter((product) => {
    if (featuredSlugs.value.has(product.slug)) return false
    if (!query) return true
    const searchable = `${getLocalizedText(product.title)} ${getLocalizedText(product.description)} ${product.slug}`.toLocaleLowerCase()
    return searchable.includes(query)
  })
})
const formatProductPrice = (amount: unknown) => {
  const value = Number(amount)
  return isPriceAvailable(amount) ? `¥${value.toFixed(2)}` : '价格暂不可用'
}
const isPriceAvailable = (amount: unknown) => Number.isFinite(Number(amount)) && Number(amount) >= 0
const posts = ref<any[]>([])
const quickBuyProduct = ref<any>(null)
const quickBuyVisible = ref(false)

const { shouldShow } = useAnnouncement()
const activeAnnouncement = ref<HomeAnnouncement | null>(null)
const announcementVisible = ref(false)

let backgroundScrollY = 0
let backgroundScrollLocked = false
let previousPageOverflow = ''
let previousBodyPosition = ''
let previousBodyTop = ''
let previousBodyLeft = ''
let previousBodyRight = ''
let previousBodyWidth = ''
watch([mobileTermsItem, desktopDetailsItem, quickBuyVisible, announcementVisible], ([terms, details, quickBuy, announcement]) => {
  const shouldLock = Boolean(terms || details || quickBuy || announcement)
  if (typeof document === 'undefined' || shouldLock === backgroundScrollLocked) return
  if (shouldLock) {
    backgroundScrollY = window.scrollY
    previousPageOverflow = document.documentElement.style.overflow
    previousBodyPosition = document.body.style.position
    previousBodyTop = document.body.style.top
    previousBodyLeft = document.body.style.left
    previousBodyRight = document.body.style.right
    previousBodyWidth = document.body.style.width
    document.documentElement.style.overflow = 'hidden'
    document.body.style.position = 'fixed'
    document.body.style.top = `-${backgroundScrollY}px`
    document.body.style.left = '0'
    document.body.style.right = '0'
    document.body.style.width = '100%'
    backgroundScrollLocked = true
  } else {
    document.documentElement.style.overflow = previousPageOverflow
    document.body.style.position = previousBodyPosition
    document.body.style.top = previousBodyTop
    document.body.style.left = previousBodyLeft
    document.body.style.right = previousBodyRight
    document.body.style.width = previousBodyWidth
    window.scrollTo(0, backgroundScrollY)
    backgroundScrollLocked = false
  }
}, { flush: 'sync' })
onBeforeUnmount(() => {
  if (!backgroundScrollLocked || typeof document === 'undefined') return
  document.documentElement.style.overflow = previousPageOverflow
  document.body.style.position = previousBodyPosition
  document.body.style.top = previousBodyTop
  document.body.style.left = previousBodyLeft
  document.body.style.right = previousBodyRight
  document.body.style.width = previousBodyWidth
  window.scrollTo(0, backgroundScrollY)
  backgroundScrollLocked = false
})

const navBuiltin = computed(() => (appStore.config?.nav_config as { builtin?: Record<string, boolean> } | undefined)?.builtin)
const blogEnabled = computed(() => navBuiltin.value?.blog !== false)
const noticeEnabled = computed(() => navBuiltin.value?.notice !== false)
const latestSectionVisible = computed(() => blogEnabled.value || noticeEnabled.value)

const openQuickBuy = (product: any) => {
  quickBuyProduct.value = product
  quickBuyVisible.value = true
}

const handleQuickBuy = (slug: string) => {
  // The main storefront advertises the platform price. A previous visit to the
  // agent center must not carry its wholesale tenant into this purchase.
  if (!['agent', 'reseller', 'ref', 'aff'].some((key) => router.currentRoute.value.query[key])) {
    sessionStorage.removeItem('dj_agent_code')
  }
  const found = products.value.find((p) => p.slug === slug)
  if (found) {
    openQuickBuy(found)
  } else {
    productAPI.detail(slug).then((res: any) => {
      if (res.data?.data) {
        openQuickBuy(res.data.data)
      } else {
        router.push(`/products/${slug}`)
      }
    }).catch(() => {
      router.push(`/products/${slug}`)
    })
  }
}

const formatDate = (dateString: string) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString()
}

const goToPost = (slug: string) => {
  router.push(`/blog/${slug}`)
}

const loadProducts = async () => {
  productsLoading.value = true
  productsLoaded.value = false
  productsLoadError.value = ''
  try {
    const response = await productAPI.list({ page: 1, page_size: 100 })
    const list = response.data?.data || []
    products.value = list
    productsLoaded.value = true

    // 动态同步后台设置的最新商品售价
    for (const item of featuredList.value) {
      const live = list.find((p: any) => p.slug === item.slug)
      if (live) {
        if (live.price_amount) {
          item.price = parseFloat(live.price_amount).toFixed(2)
        }
        if (live.title) {
          const localizedTitle = getLocalizedText(live.title)
          if (localizedTitle) {
            item.name = localizedTitle
          }
        }
      }
    }
  } catch (error) {
    productsLoadError.value = '商品价格加载失败，请刷新页面后重试。'
    console.error('Failed to load products:', error)
  } finally {
    productsLoading.value = false
  }
}

const loadLatestPosts = async () => {
  if (!latestSectionVisible.value) return
  try {
    const params: Record<string, unknown> = { page: 1, page_size: 3 }
    if (blogEnabled.value && !noticeEnabled.value) params.type = 'blog'
    if (!blogEnabled.value && noticeEnabled.value) params.type = 'notice'
    const response = await postAPI.list(params)
    posts.value = response.data?.data || []
  } catch (error) {
    console.error('Failed to load posts:', error)
  }
}

const showAnnouncementIfNeeded = () => {
  const announcement = appStore.config?.announcement as HomeAnnouncement | undefined
  if (announcement && shouldShow(announcement)) {
    activeAnnouncement.value = announcement
    announcementVisible.value = true
  }
}

onMounted(async () => {
  await appStore.loadConfig()
  await Promise.all([loadProducts(), loadLatestPosts()])
  showAnnouncementIfNeeded()
})
</script>

<style scoped>
.scale-102 {
  transform: scale(1.02);
}
.scale-103 {
  transform: scale(1.03);
}
</style>
