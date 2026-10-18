<template>
  <div class="min-h-screen bg-[var(--ui-bg-page)] py-12 px-4">
    <div class="max-w-6xl mx-auto space-y-8">
      <!-- Header -->
      <div class="text-center">
        <h1 class="text-4xl font-bold text-[var(--ui-text-primary)] mb-2">
          2024 UI Redesign - 配色系统
        </h1>
        <p class="text-[var(--ui-text-secondary)]">
          田园绿 + 珍珠灰配色方案演示
        </p>
      </div>

      <!-- Primary Colors -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">主色调 - 田园绿</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-accent)] rounded-lg flex items-center justify-center text-white font-medium">
              Primary
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">--ui-accent: #52A883</p>
          </div>
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-accent-hover)] rounded-lg flex items-center justify-center text-white font-medium">
              Primary Hover
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">--ui-accent-hover: #459573</p>
          </div>
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-accent-soft)] rounded-lg flex items-center justify-center text-[var(--ui-accent)] font-medium border border-[var(--ui-accent)]">
              Primary Soft
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">--ui-accent-soft: #EBF7F2</p>
          </div>
        </div>
      </section>

      <!-- Backgrounds -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">背景色 - 珍珠灰系</h2>
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-bg-page)] rounded-lg flex items-center justify-center text-[var(--ui-text-primary)] font-medium border border-[var(--ui-border)]">
              Page
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">#F8F7F4</p>
          </div>
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-bg-elevated)] rounded-lg flex items-center justify-center text-[var(--ui-text-primary)] font-medium border border-[var(--ui-border)]">
              Elevated
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">#FFFFFF</p>
          </div>
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-bg-soft)] rounded-lg flex items-center justify-center text-[var(--ui-text-primary)] font-medium border border-[var(--ui-border)]">
              Soft
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">#F0EFE9</p>
          </div>
          <div class="space-y-2">
            <div class="h-24 bg-[var(--ui-bg-muted)] rounded-lg flex items-center justify-center text-[var(--ui-text-primary)] font-medium border border-[var(--ui-border)]">
              Muted
            </div>
            <p class="text-sm text-[var(--ui-text-secondary)]">#E8E6D9</p>
          </div>
        </div>
      </section>

      <!-- Semantic Colors -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">功能色</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <div class="h-20 bg-[var(--ui-success-soft)] rounded-lg flex items-center justify-center border-2 border-[var(--ui-success)]">
              <span class="text-[var(--ui-success)] font-semibold">成功 Success</span>
            </div>
          </div>
          <div class="space-y-2">
            <div class="h-20 bg-[var(--ui-warning-soft)] rounded-lg flex items-center justify-center border-2 border-[var(--ui-warning)]">
              <span class="text-[var(--ui-warning)] font-semibold">警告 Warning</span>
            </div>
          </div>
          <div class="space-y-2">
            <div class="h-20 bg-[var(--ui-danger-soft)] rounded-lg flex items-center justify-center border-2 border-[var(--ui-danger)]">
              <span class="text-[var(--ui-danger)] font-semibold">错误 Error</span>
            </div>
          </div>
          <div class="space-y-2">
            <div class="h-20 bg-[var(--ui-info-soft)] rounded-lg flex items-center justify-center border-2 border-[var(--ui-info)]">
              <span class="text-[var(--ui-info)] font-semibold">信息 Info</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Buttons -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">按钮样式</h2>
        <div class="flex flex-wrap gap-4">
          <button class="px-6 py-3 bg-[var(--ui-accent)] hover:bg-[var(--ui-accent-hover)] text-white rounded-lg font-medium transition-colors">
            Primary Button
          </button>
          <button class="px-6 py-3 bg-[var(--ui-bg-soft)] hover:bg-[var(--ui-bg-muted)] text-[var(--ui-text-primary)] rounded-lg font-medium transition-colors">
            Secondary Button
          </button>
          <button class="px-6 py-3 border-2 border-[var(--ui-accent)] text-[var(--ui-accent)] hover:bg-[var(--ui-accent-soft)] rounded-lg font-medium transition-colors">
            Outline Button
          </button>
          <button class="px-6 py-3 bg-[var(--ui-danger)] hover:bg-[#C75B5B] text-white rounded-lg font-medium transition-colors">
            Danger Button
          </button>
        </div>
      </section>

      <!-- Cards -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">卡片样式</h2>
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div class="bg-[var(--ui-bg-elevated)] border border-[var(--ui-border)] rounded-lg p-4 shadow-[var(--ui-shadow-card)]">
            <h3 class="text-lg font-semibold text-[var(--ui-text-primary)] mb-2">标准卡片</h3>
            <p class="text-[var(--ui-text-secondary)]">白色背景，柔和边框和阴影</p>
            <div class="mt-4">
              <span class="inline-block px-3 py-1 bg-[var(--ui-accent)] text-white text-sm rounded-full">
                标签
              </span>
            </div>
          </div>

          <div class="bg-[var(--ui-accent-soft)] border border-[var(--ui-accent)] rounded-lg p-4">
            <h3 class="text-lg font-semibold text-[var(--ui-accent)] mb-2">强调卡片</h3>
            <p class="text-[var(--ui-text-secondary)]">浅绿背景，田园绿边框</p>
            <div class="mt-4">
              <span class="inline-block px-3 py-1 bg-[var(--ui-accent)] text-white text-sm rounded-full">
                代理价
              </span>
            </div>
          </div>

          <div class="bg-[var(--ui-bg-soft)] border border-[var(--ui-border)] rounded-lg p-4">
            <h3 class="text-lg font-semibold text-[var(--ui-text-primary)] mb-2">次级卡片</h3>
            <p class="text-[var(--ui-text-secondary)]">珍珠灰背景，柔和视觉</p>
            <div class="mt-4">
              <span class="inline-block px-3 py-1 bg-[var(--ui-bg-muted)] text-[var(--ui-text-secondary)] text-sm rounded-full">
                辅助信息
              </span>
            </div>
          </div>
        </div>
      </section>

      <!-- Typography -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">文字层级</h2>
        <div class="space-y-3">
          <div>
            <span class="text-[var(--ui-text-primary)] font-bold text-2xl">主要文字 (Primary)</span>
            <span class="text-sm text-[var(--ui-text-muted)] ml-4">#1A1A1A</span>
          </div>
          <div>
            <span class="text-[var(--ui-text-secondary)] font-medium text-xl">次要文字 (Secondary)</span>
            <span class="text-sm text-[var(--ui-text-muted)] ml-4">#555555</span>
          </div>
          <div>
            <span class="text-[var(--ui-text-muted)] text-base">辅助文字 (Muted)</span>
            <span class="text-sm text-[var(--ui-text-muted)] ml-4">#888888</span>
          </div>
        </div>
      </section>

      <!-- Demo Content -->
      <section class="bg-[var(--ui-bg-elevated)] rounded-lg p-6 shadow-[var(--ui-shadow-card)]">
        <h2 class="text-2xl font-semibold text-[var(--ui-text-primary)] mb-4">商品卡片示例</h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <!-- Product Card -->
          <div class="bg-[var(--ui-bg-elevated)] border border-[var(--ui-border)] rounded-lg overflow-hidden shadow-[var(--ui-shadow-card)] hover:shadow-[var(--ui-shadow-2)] transition-shadow">
            <div class="h-32 bg-gradient-to-br from-[var(--ui-accent-soft)] to-[var(--ui-bg-soft)] flex items-center justify-center">
              <span class="text-4xl">🤖</span>
            </div>
            <div class="p-4">
              <h3 class="text-lg font-semibold text-[var(--ui-text-primary)] mb-2">
                GPT-4 Plus 订阅 - 1个月
              </h3>
              <p class="text-sm text-[var(--ui-text-secondary)] mb-3">
                包含GPT-4、DALL·E 3、Advanced Data Analysis等高级功能
              </p>
              <div class="flex items-baseline gap-2 mb-3">
                <span class="text-2xl font-bold text-[var(--ui-accent)]">¥138</span>
                <span class="text-sm text-[var(--ui-text-muted)] line-through">¥168</span>
                <span class="ml-auto px-2 py-1 bg-[var(--ui-accent)] text-white text-xs rounded">
                  代理价
                </span>
              </div>
              <button class="w-full py-2 bg-[var(--ui-accent)] hover:bg-[var(--ui-accent-hover)] text-white rounded-lg font-medium transition-colors">
                立即购买
              </button>
            </div>
          </div>

          <!-- Order Status Card -->
          <div class="bg-[var(--ui-bg-elevated)] border border-[var(--ui-border)] rounded-lg p-4 shadow-[var(--ui-shadow-card)]">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-semibold text-[var(--ui-text-primary)]">订单状态</h3>
              <span class="px-3 py-1 bg-[var(--ui-success-soft)] text-[var(--ui-success)] text-sm rounded-full border border-[var(--ui-success)]">
                已完成
              </span>
            </div>
            <div class="space-y-3">
              <div class="flex justify-between text-sm">
                <span class="text-[var(--ui-text-secondary)]">订单号</span>
                <span class="text-[var(--ui-text-primary)] font-mono">AI202410170001</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-[var(--ui-text-secondary)]">商品</span>
                <span class="text-[var(--ui-text-primary)]">GPT-4 Plus 1个月</span>
              </div>
              <div class="flex justify-between text-sm">
                <span class="text-[var(--ui-text-secondary)]">金额</span>
                <span class="text-lg font-bold text-[var(--ui-accent)]">¥138.00</span>
              </div>
              <div class="pt-3 border-t border-[var(--ui-border)]">
                <div class="bg-[var(--ui-accent-soft)] border border-[var(--ui-accent)] rounded-lg p-3">
                  <p class="text-sm text-[var(--ui-text-primary)] font-medium mb-1">Session已提交</p>
                  <p class="text-xs text-[var(--ui-text-secondary)]">您的Session已成功提交到账户</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

    </div>
  </div>
</template>

<script setup lang="ts">
// Color demo page - 展示新配色系统
</script>
